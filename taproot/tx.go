package taproot

import (
	"github.com/quocky/taproot-asset/taproot/onchain"
	"log"
)

type GetUnspentAssetsByIdResult struct {
}

// prepareTx create simple minting asset template
func (t *Taproot) prepareTx(
	utxosResult *onchain.UTXOResult,
	unspentAssets *GetUnspentAssetsByIdResult,
	defaultOutputAmount int32,
	receivers []*onchain.Receiver,
	isMint bool,
) (*onchain.TxIncludeOutInternalKey, error) {
	unspentAssetsOnchains, err := makeUnspentAssetsByIdResult(unspentAssets)
	if err != nil {
		log.Println("makeUnspentAssetsByIdResult err", err)

		return nil, err
	}

	txMaker, err := t.btcClient.NewTxMaker(
		utxosResult, unspentAssetsOnchains,
		defaultOutputAmount, receivers,
	)
	if err != nil {
		log.Println("[prepare tx] cannot create TxMaker, ", err)
		return nil, err
	}

	err = txMaker.CreateTemplateTx()
	if err != nil {
		log.Println("[prepare tx] cannot create templateTx, ", err)
		return nil, err
	}

	//err = txMaker.AddRevealData(isMint)
	//if err != nil {
	//	log.Println("[prepare tx] cannot add reveal data, ", err)
	//	return nil, err
	//}

	if err := txMaker.SignTaprootInput(t.wif.PrivKey); err != nil {
		log.Println("txMaker.SignTaprootInput err", err)

		return nil, err
	}

	finalTx, err := t.btcClient.SignRawTx(txMaker.Tx)
	if err != nil {
		log.Println("t.onchain.SignRawTx(txMaker.Tx) fail", err)

		return nil, err
	}

	return &onchain.TxIncludeOutInternalKey{
		Tx:              finalTx,
		OutInternalKeys: txMaker.OutputInternalKeys,
	}, nil
}

func makeUnspentAssetsByIdResult(
	unspentAssets *GetUnspentAssetsByIdResult,
) ([]*onchain.UnspentAssetsByIdResult, error) {
	//if unspentAssets == nil {
	//	return nil, nil
	//}
	//
	//unspentAssetsOnchains := make([]*onchain.UnspentAssetsByIdResult, len(unspentAssets.UnspentOutpoints))
	//for id, unspentOutpoint := range unspentAssets.UnspentOutpoints {
	//
	//	outpoint, err := wire.NewOutPointFromString(unspentOutpoint.Outpoint)
	//	if err != nil {
	//		log.Println("[makeUnspentAssetsByIdResult] wire.NewOutPointFromString, err", err)
	//
	//		return nil, err
	//	}
	//
	//	unspentAssetsOnchains[id] = &onchain.UnspentAssetsByIdResult{
	//		Outpoint:         outpoint,
	//		AmtSats:          int64(unspentOutpoint.AmtSats),
	//		ScriptOutput:     unspentOutpoint.ScriptOutput,
	//		InternalKey:      unspentOutpoint.InternalKey,
	//		TaprootAssetRoot: unspentOutpoint.TaprootAssetRoot,
	//	}
	//}
	//
	//return unspentAssetsOnchains, nil

	return nil, nil
}
