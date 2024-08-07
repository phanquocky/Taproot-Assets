package onchain

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
)

type TxIncludeOutPubKey struct {
	Tx         *wire.MsgTx
	OutPubKeys map[int32]asset.SerializedKey
}

type UnspentAssetsByIdResult struct {
	Outpoint         *wire.OutPoint
	AmtSats          int64
	ScriptOutput     []byte
	InternalKey      []byte
	TaprootAssetRoot []byte
}

// TxMaker struct using create format onchain tx
type TxMaker struct {
	UTXOs          []*UnspentTXOut
	unspentAssets  []*UnspentAssetsByIdResult
	senderAddress  btcutil.Address
	btcOutputInfos []*BtcOutputInfo
	fee            btcutil.Amount

	Tx            *wire.MsgTx
	OutputPubKeys map[int32]asset.SerializedKey
}

func (c *Client) NewTxMaker(
	UTXOs []*UnspentTXOut,
	unspentAssets []*UnspentAssetsByIdResult,
	outputInfos []*BtcOutputInfo,
	senderBtcAddress btcutil.Address,
	fee btcutil.Amount,
) (*TxMaker, error) {

	return &TxMaker{
		UTXOs:          UTXOs,
		unspentAssets:  unspentAssets,
		senderAddress:  senderBtcAddress,
		btcOutputInfos: outputInfos,
		fee:            fee,
		Tx:             nil,
		OutputPubKeys:  make(map[int32]asset.SerializedKey),
	}, nil
}

func (t *TxMaker) CreateTemplateTx() error {
	tx := wire.NewMsgTx(2)

	inputAmount := btcutil.Amount(0)
	outputAmount := btcutil.Amount(0)

	if t.unspentAssets != nil { // TODO:
		for _, unspent := range t.unspentAssets {
			fmt.Printf("unspent: %v\n", unspent)
			inputAmount += btcutil.Amount(unspent.AmtSats)
			fmt.Printf("unspend outpoint %v\n", unspent.Outpoint)
			tx.AddTxIn(wire.NewTxIn(unspent.Outpoint, nil, nil))

		}
	}

	for _, u := range t.UTXOs {
		inputAmount += u.Amount
		tx.AddTxIn(wire.NewTxIn(u.Outpoint, nil, nil))
	}

	for i, output := range t.btcOutputInfos {
		outputAmount += btcutil.Amount(output.SatAmount)
		pkScript, err := txscript.PayToAddrScript(output.AddrResult.Address)
		if err != nil {
			return err
		}

		t.OutputPubKeys[int32(i)] = output.AddrResult.PubKey

		tx.AddTxOut(wire.NewTxOut(int64(output.SatAmount), pkScript))
	}

	if outputAmount > inputAmount {
		return errors.New("output amount is greater than input amount")
	}

	if inputAmount-outputAmount-t.fee > 0 {
		pkScript, err := txscript.PayToAddrScript(t.senderAddress)
		if err != nil {
			return err
		}
		tx.AddTxOut(wire.NewTxOut(int64(inputAmount-outputAmount-t.fee), pkScript))
	}

	t.Tx = tx
	return nil
}

// AddRevealData function add reveal data to last output in onchain tx
func (t *TxMaker) AddRevealData(isMint bool) error {
	data := make([]byte, 0)

	// map assetID = [][outputindex, asset amount]
	AssetIDmap := make(map[[32]byte][][2]uint32)

	//for i, receiver := range t.btcOutputInfos {
	//	assetID := receiver.OutputAsset.ID()
	//	AssetIDmap[assetID] = append(AssetIDmap[assetID], [2]uint32{uint32(i), uint32(receiver.OutputAsset.Amount)})
	//}

	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddOp(txscript.OP_RETURN)

	if isMint {
		scriptBuilder.AddData([]byte{'t', 'a', 'r', 'o'})
		scriptBuilder.AddOp(txscript.OP_0)
	}

	for id, outs := range AssetIDmap {
		data = append(data, id[:]...)

		for _, out := range outs {
			a := make([]byte, 8)

			binary.LittleEndian.PutUint32(a, out[0])
			binary.LittleEndian.PutUint32(a[4:], out[1])
			data = append(data, a[:]...)
		}
	}

	if len(data) > 9900 || len(data) <= 0 {
		return errors.New("len data invalid")
	}

	for i := 0; i < (len(data)+519)/520; i++ {
		curData := data[i*520 : min((i+1)*520, len(data))]

		scriptBuilder.AddData(curData)
		scriptBuilder.AddOp(txscript.OP_0)
	}

	scriptBuilderByte, err := scriptBuilder.Script()
	if err != nil {
		log.Println("[CreateRevealData] cannot scriptbuilder to script bytes, err ", err)
		return err
	}

	t.Tx.AddTxOut(&wire.TxOut{
		Value:    0,
		PkScript: scriptBuilderByte[0 : len(scriptBuilderByte)-1],
	})

	return nil
}
func (t *TxMaker) createPrevOutFetchers() *txscript.MultiPrevOutFetcher {
	prevOutFetchers := txscript.NewMultiPrevOutFetcher(nil)

	for _, u := range t.UTXOs {
		prevOutFetchers.AddPrevOut(*u.Outpoint, wire.NewTxOut(int64(u.Amount), u.LockScript))
	}

	// TODO:
	for _, unspent := range t.unspentAssets {
		fmt.Println("scriptOutput: ", unspent.ScriptOutput)
		prevOutFetchers.AddPrevOut(*unspent.Outpoint, wire.NewTxOut(int64(unspent.AmtSats), unspent.ScriptOutput))
	}

	fmt.Println("prevOutFetchers: ", prevOutFetchers)

	return prevOutFetchers

}

// SignTaprootInput function sign only taproot input in onchain transaction
func (t *TxMaker) SignTaprootInput(privKey *btcec.PrivateKey) error {
	prevOutFetchers := t.createPrevOutFetchers()

	if t.unspentAssets != nil {

		for index, unspent := range t.unspentAssets {

			sigHashes := txscript.NewTxSigHashes(t.Tx, prevOutFetchers)
			tapScriptRootHash := unspent.TaprootAssetRoot

			sig, err := txscript.RawTxInTaprootSignature(
				t.Tx, sigHashes, index,
				unspent.AmtSats,
				unspent.InternalKey,
				tapScriptRootHash,
				txscript.SigHashDefault,
				privKey,
			)
			if err != nil {
				fmt.Println("error sign taproot input: ", err)
				return err
			}

			t.Tx.TxIn[index].Witness = wire.TxWitness{sig}
		}

	}

	return nil
}
