package taproot

import (
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

type OutPoint struct {
	Outpoint         string
	AmtSats          int64
	ScriptOutput     []byte
	InternalKey      []byte
	TaprootAssetRoot []byte
}

type GetUnspentAssetsByIdResult struct {
	UnspentOutpoints []*OutPoint
}

func (t *Taproot) createTxOnChain(
	UTXOs []*onchain.UnspentTXOut,
	unspentAssets *GetUnspentAssetsByIdResult, // TODO:
	outputInfos []*onchain.BtcOutputInfo,
	fee btcutil.Amount,
	isMint bool,
) (*onchain.TxIncludeOutPubKey, error) {
	unspentAssetsOnChains, err := makeUnspentAssetsByIdResult(unspentAssets)
	if err != nil {
		return nil, err
	}

	senderBtcAddr, err := t.btcClient.GetSenderAddress()
	if err != nil {
		return nil, err
	}

	txMaker, err := t.btcClient.NewTxMaker(
		UTXOs, unspentAssetsOnChains,
		outputInfos,
		senderBtcAddr,
		fee,
	)
	if err != nil {
		return nil, err
	}

	err = txMaker.CreateTemplateTx()
	if err != nil {
		return nil, err
	}

	//err = txMaker.AddRevealData(isMint)
	//if err != nil {
	//	log.Println("[prepare tx] cannot add reveal data, ", err)
	//	return nil, err
	//}

	if err := txMaker.SignTaprootInput(t.wif.PrivKey); err != nil {
		return nil, err
	}

	log.Println("[createTxOnChain] txMaker.Tx: ", txMaker.Tx)
	finalTx, err := t.btcClient.SignRawTx(txMaker.Tx)
	if err != nil {
		return nil, err
	}
	log.Println("[createTxOnChain] finalTx: ", finalTx)

	return &onchain.TxIncludeOutPubKey{
		Tx:         finalTx,
		OutPubKeys: txMaker.OutputPubKeys,
	}, nil
}

func makeUnspentAssetsByIdResult(
	unspentAssets *GetUnspentAssetsByIdResult,
) ([]*onchain.UnspentAssetsByIdResult, error) {
	if unspentAssets == nil {
		return nil, nil
	}

	unspentAssetsOnchains := make([]*onchain.UnspentAssetsByIdResult, len(unspentAssets.UnspentOutpoints))
	for id, unspentOutpoint := range unspentAssets.UnspentOutpoints {

		outpoint, err := wire.NewOutPointFromString(unspentOutpoint.Outpoint)
		if err != nil {
			log.Println("[makeUnspentAssetsByIdResult] wire.NewOutPointFromString, err", err)

			return nil, err
		}

		unspentAssetsOnchains[id] = &onchain.UnspentAssetsByIdResult{
			Outpoint:         outpoint,
			AmtSats:          int64(unspentOutpoint.AmtSats),
			ScriptOutput:     unspentOutpoint.ScriptOutput,
			InternalKey:      unspentOutpoint.InternalKey,
			TaprootAssetRoot: unspentOutpoint.TaprootAssetRoot,
		}
	}

	return unspentAssetsOnchains, nil
}
