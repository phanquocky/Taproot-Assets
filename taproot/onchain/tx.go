package onchain

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/address"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"log"
)

type TxIncludeOutInternalKey struct {
	Tx              *wire.MsgTx
	OutInternalKeys map[int]asset.SerializedKey
}

// Receiver is struct contain all information of a btc output
type Receiver struct {
	AddrResult  *address.AddrResult
	OutputAsset []*asset.Asset
}

func (r *Receiver) GetAddrResult() *address.AddrResult {
	if r == nil {
		return nil
	}

	return r.AddrResult
}

func (r *Receiver) GetOutputAsset() []*asset.Asset {
	if r == nil {
		return nil
	}

	return r.OutputAsset
}

func NewReceiver(addrResult *address.AddrResult, outputAsset ...*asset.Asset) *Receiver {
	return &Receiver{
		AddrResult:  addrResult,
		OutputAsset: outputAsset,
	}
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
	utxos               *UTXOResult
	unspentAssets       []*UnspentAssetsByIdResult
	senderAddress       btcutil.Address
	defaultOutputAmount int32
	receivers           []*Receiver

	Tx                 *wire.MsgTx
	OutputInternalKeys map[int]asset.SerializedKey
}

// NewTxMaker func create new TxMaker struct
func (c *Client) NewTxMaker(
	utxos *UTXOResult,
	unspentAssets []*UnspentAssetsByIdResult,
	defaultOutputAmount int32,
	receivers []*Receiver,
) (*TxMaker, error) {

	senderAddrBtc, err := btcutil.DecodeAddress(c.networkConfig.SenderAddress, c.networkConfig.ParamsObject)
	if err != nil {
		panic("your sender address is invalid, " + c.networkConfig.SenderAddress + err.Error())
		return nil, err
	}

	return &TxMaker{
		utxos:               utxos,
		unspentAssets:       unspentAssets,
		senderAddress:       senderAddrBtc,
		defaultOutputAmount: defaultOutputAmount,
		receivers:           receivers,
		Tx:                  nil,
		OutputInternalKeys:  make(map[int]asset.SerializedKey),
	}, nil
}

// CreateTemplateTx function assign inputs, and outputs to tx, return amount btc
// with no reveal data,
func (t *TxMaker) CreateTemplateTx() error {
	tx := wire.NewMsgTx(2)

	if t.unspentAssets != nil {
		for _, unspent := range t.unspentAssets {
			tx.AddTxIn(wire.NewTxIn(unspent.Outpoint, nil, nil))
		}
	}

	for _, utxo := range t.utxos.UTXOs {
		tx.AddTxIn(wire.NewTxIn(utxo, nil, nil))
	}

	for i, receiver := range t.receivers {
		pkscript, err := txscript.PayToAddrScript(receiver.AddrResult.Address)
		if err != nil {
			log.Println("cannot create pkscript from tapaddress: ", receiver)
			return err
		}

		t.OutputInternalKeys[i] = receiver.AddrResult.Pubkey

		tx.AddTxOut(wire.NewTxOut(int64(t.defaultOutputAmount), pkscript))
	}

	if t.utxos.TotalAmount > t.utxos.ActualAmount {
		pkscript, err := txscript.PayToAddrScript(t.senderAddress)
		if err != nil {
			log.Println("cannot create pkscript from tapaddress: ", t.senderAddress)
			return err
		}
		tx.AddTxOut(wire.NewTxOut(int64(t.utxos.TotalAmount-t.utxos.ActualAmount), pkscript))
	}

	t.Tx = tx
	return nil
}

// AddRevealData function add reveal data to last output in onchain tx
func (t *TxMaker) AddRevealData(isMint bool) error {
	data := make([]byte, 0)

	// map assetID = [][outputindex, asset amount]
	AssetIDmap := make(map[[32]byte][][2]uint32)

	//for i, receiver := range t.receivers {
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

// SignTaprootInput function sign only taproot input in onchain transaction
func (t *TxMaker) SignTaprootInput(privKey *btcec.PrivateKey) error {

	if t.unspentAssets != nil {

		var inputFetchers = t.utxos.InputFetcher
		// this for loop create unspent
		for _, unspent := range t.unspentAssets {
			inputFetchers.AddPrevOut(*unspent.Outpoint, wire.NewTxOut(int64(unspent.AmtSats), unspent.ScriptOutput))
		}

		//sign taproot asset input
		for index, unspent := range t.unspentAssets {

			sigHashes := txscript.NewTxSigHashes(t.Tx, inputFetchers)
			tapscriptRootHash := unspent.TaprootAssetRoot

			sig, err := txscript.RawTxInTaprootSignature(
				t.Tx, sigHashes, index,
				unspent.AmtSats,
				unspent.InternalKey,
				tapscriptRootHash,
				txscript.SigHashSingle,
				privKey,
			)
			if err != nil {
				fmt.Println("[signTaprootInput] txscript.RawTxInTaprootSignature, err ", err)
				return err
			}

			t.Tx.TxIn[index].Witness = wire.TxWitness{sig}
		}

	}

	return nil
}
