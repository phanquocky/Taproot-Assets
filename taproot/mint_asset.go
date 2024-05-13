package taproot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/http_model/mint"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

func (t *Taproot) MintAsset(ctx context.Context, assetNames []string, assetAmounts []int32) error {
	err := preCheckAssets(assetNames, assetAmounts)
	if err != nil {
		return err
	}

	log.Println("[Mint Asset] Precheck assets success!")

	var (
		expectBtcAmount = int32(DEFAULT_OUTPUT_AMOUNT + DEFAULT_FEE)
		userPubKey      = asset.ToSerialized(t.wif.PrivKey.PubKey())
	)

	btcUTXOs, err := t.btcClient.ListUTXOs()
	if err != nil {
		return err
	}

	if len(btcUTXOs) == 0 {
		return errors.New("utxos is empty")
	}

	bestUTXOs, err := chooseBestUTXOs(btcUTXOs, expectBtcAmount)
	if err != nil {
		return err
	}

	log.Println("[Mint Asset] Choose best utxos success!")

	firstPrevOut := bestUTXOs[0].Outpoint
	mintAssets := genAssets(assetNames, assetAmounts, firstPrevOut, userPubKey)

	log.Println("[Mint Asset] Generate assets success!", mintAssets)

	assetCommitments, err := genAssetCommitments(ctx, mintAssets)
	if err != nil {
		return err
	}

	log.Println("[Mint Asset] Generate asset commitments success!")

	tapCommitment, err := commitment.NewTapCommitment(assetCommitments...)
	if err != nil {
		return err
	}

	log.Println("[Mint Asset] Generate tap commitment success!")

	mintTapAddress, err := t.addressMaker.CreateTapAddr(userPubKey, tapCommitment)
	if err != nil {
		return err
	}

	log.Println("[Mint Asset] Generate tap address success!")

	btcOutputInfos := []*onchain.BtcOutputInfo{
		onchain.NewBtcOutputInfo(mintTapAddress, DEFAULT_OUTPUT_AMOUNT, mintAssets...),
	}

	txIncludeOutPubKey, err := t.createTxOnChain(bestUTXOs, nil,
		btcOutputInfos, btcutil.Amount(DEFAULT_FEE), true)
	if err != nil {
		return err
	}

	log.Println("[Mint Asset] Create tx on chain success!")

	mintProof, err := createMintProof(
		txIncludeOutPubKey,
		DEFAULT_MINTING_OUTPUT_INDEX,
		tapCommitment,
	)

	log.Println("[Mint Asset] Create mint proof success!")

	data := mint.MintAssetReq{
		AmountSats:        expectBtcAmount,
		TapScriptRootHash: mintTapAddress.TapScriptRootHash,
		MintProof:         mintProof,
	}

	postResp, err := t.httpClient.R().SetBody(data).Post(os.Getenv("SERVER_BASE_URL") + "/mint-asset")
	if err != nil {
		log.Println("c.httpClient.R().SetBody(data).Post(\"/mint-asset\")", err)

		return err
	}

	log.Println("[Mint Asset] Post mint asset success!", postResp)

	return nil
}

func chooseBestUTXOs(utxos []*onchain.UnspentTXOut, expectAmount int32) ([]*onchain.UnspentTXOut, error) {
	var (
		bestUTXOs                      = make([]*onchain.UnspentTXOut, 0)
		expectSatAmount                = btcutil.Amount(expectAmount)
		totalAmount     btcutil.Amount = 0
	)
	for _, utxo := range utxos {
		bestUTXOs = append(bestUTXOs, utxo)
		totalAmount += utxo.Amount

		if totalAmount >= expectSatAmount {
			return bestUTXOs, nil
		}
	}

	return nil, errors.New("not enough utxos")
}

func preCheckAssets(assetNames []string, assetAmounts []int32) error {
	if len(assetNames) != len(assetAmounts) {
		return errors.New("len assetNames and amount is different")
	}

	for idx, assetName := range assetNames {
		assetAmount := assetAmounts[idx]

		if len(assetName) < 6 {
			return fmt.Errorf("expected len your asset assetName is greater than or equal 6 (but len = %d) ", len(assetName))
		}

		if assetAmount == 0 {
			return fmt.Errorf("expected assetAmount greater than 0 (but assetAmount = %d)", len(assetName))
		}
	}

	return nil
}

func genAssets(assetNames []string, assetAmounts []int32, prevOut *wire.OutPoint, userPubKey asset.SerializedKey) []*asset.Asset {

	var assets = make([]*asset.Asset, len(assetNames))

	for idx, assetName := range assetNames {
		assetAmount := assetAmounts[idx]

		log.Printf("[Mint Asset] mintint asset! assetName : %v, assetAmount : %v ! \n", assetName, assetAmount)

		genesis := asset.NewGenesis(*prevOut, assetName, DEFAULT_MINTING_OUTPUT_INDEX)
		assets[idx] = asset.NewAsset(genesis, assetAmount, userPubKey, nil)
	}

	return assets
}

func genAssetCommitments(ctx context.Context, assets []*asset.Asset) ([]*commitment.AssetCommitment, error) {
	var assetCommitments = make([]*commitment.AssetCommitment, len(assets))

	for idx, a := range assets {
		assetCommitment, err := commitment.NewAssetCommitment(ctx, a)
		if err != nil {
			return nil, err
		}

		assetCommitments[idx] = assetCommitment
	}

	return assetCommitments, nil
}

func createMintProof(
	txIncludeOutPubKey *onchain.TxIncludeOutPubKey,
	outputIndex int32,
	tapCommitment *commitment.TapCommitment,
) (proof.AssetProofs, error) {
	baseProof := &proof.MintParams{
		BaseProofParams: proof.BaseProofParams{
			Tx:            txIncludeOutPubKey.Tx,
			OutputIndex:   outputIndex,
			InternalKey:   txIncludeOutPubKey.OutPubKeys[outputIndex],
			TapCommitment: tapCommitment,
		},
		GenesisPoint: txIncludeOutPubKey.Tx.TxIn[0].PreviousOutPoint,
	}

	err := baseProof.BaseProofParams.AddExclusionProofs(
		txIncludeOutPubKey,
		func(idx int32) bool {
			return idx == outputIndex
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to add exclusion proofs: "+
			"%w", err)
	}

	mintProofs, err := proof.NewMintingBlobs(baseProof)
	if err != nil {
		return nil, fmt.Errorf("unable to construct minting "+
			"proofs: %w", err)
	}

	return mintProofs, nil
}
