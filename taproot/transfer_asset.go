package taproot

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/wire"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

func (t *Taproot) TransferAsset(receiverPubKey asset.SerializedKey, assetId string, amount int32) error {
	fmt.Printf("Transfer asset: asset id: %s, amount: %d, receiverPubKey: %s \n", assetId, amount, hex.EncodeToString(receiverPubKey[:]))
	ctx := context.Background()

	var (
		expectedAmount = int32(2*DEFAULT_OUTPUT_AMOUNT + DEFAULT_FEE)
	)

	UTXOs, err := t.btcClient.ListUTXOs()
	if err != nil {
		return err
	}

	bestUTXOs, err := chooseBestUTXOs(UTXOs, expectedAmount)
	if err != nil {
		return err
	}

	assetUTXOs, err := t.GetAssetUTXOs(ctx, assetId, amount)
	if err != nil {
		return err
	}

	log.Println("bestUTXOs", bestUTXOs)
	log.Println("assetUTXOs", assetUTXOs)

	//assetGenOutpoint, err := wire.NewOutPointFromString(assetUTXOs.GenesisPoint.PrevOut)
	//if err != nil {
	//	return err
	//}
	//
	//transferAsset := prepareAssets(assetGenOutpoint, assetUTXOs.GenesisAsset.AssetName, []int32{amount}, []asset.SerializedKey{receiverPubKey})
	//
	//returnAsset, err := t.createReturnAsset(assetGenOutpoint, assetUTXOs.GenesisAsset.AssetName, assetUTXOs, transferAsset)
	//if err != nil {
	//	return err
	//}
	//
	//btcOutputInfos, _, err := t.prepareBtcOutputs(ctx, assetUTXOs, transferAsset, returnAsset)
	//if err != nil {
	//	log.Println("t.createTransferAddresses(ctx, unspentAssets, transferAsset),  err ", err)
	//	return err
	//}
	//
	//txIncludeOutPubKey, err := t.createTxOnChain(bestUTXOs, nil,
	//	btcOutputInfos, btcutil.Amount(DEFAULT_FEE), true)
	//if err != nil {
	//	return err
	//}
	//
	//files, err := createFiles(
	//	assetUTXOs.InputFilesBytes,
	//	btcOutputInfos,
	//	txIncludeOutPubKey.Tx,
	//)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Println("files: ", files)
	//
	//data := transfer.TransferReq{
	//	GenesisAsset:     &assetUTXOs.GenesisAsset,
	//	AnchorTx:         txIncludeOutPubKey.Tx,
	//	AmtSats:          DEFAULT_OUTPUT_AMOUNT,
	//	BtcOutputInfos:   btcOutputInfos,
	//	UnspentOutpoints: assetUTXOs.UnspentOutpoints,
	//	Files:            files,
	//}

	if err != nil {
		return err
	}

	return nil
}

func createFiles(
	inputFilesBytes [][]byte,
	btcOutputInfos []*onchain.BtcOutputInfo,
	tx *wire.MsgTx,
) ([]*proof.File, error) {
	curFiles := make([]*proof.File, len(btcOutputInfos))

	for i := range btcOutputInfos {
		exclusionProofs, err := makeExclusionProofs(i, btcOutputInfos)
		if err != nil {
			log.Println("[createFiles] makeExclusionProofs fail", err)

			return nil, err
		}

		curFile, _, err := proof.AppendTransition(inputFilesBytes, makeLocatorTransitionParams(
			i, DEFAULT_RETURN_OUTPUT_INDEX,
			tx, btcOutputInfos,
			exclusionProofs,
		))
		if err != nil {
			log.Println("proof.AppendTransition", err)

			return nil, err
		}

		curFiles[i] = curFile
	}

	return curFiles, nil
}

func makeLocatorTransitionParams(
	i, outIndex int,
	tx *wire.MsgTx,
	btcOutputInfos []*onchain.BtcOutputInfo,
	exclusionProofs []*proof.TaprootProof,
) *proof.TransitionParams {
	return &proof.TransitionParams{
		BaseProofParams: proof.BaseProofParams{
			Tx:              tx,
			OutputIndex:     int32(i),
			InternalKey:     btcOutputInfos[i].AddrResult.PubKey,
			TapCommitment:   btcOutputInfos[i].AddrResult.TapCommitment,
			ExclusionProofs: exclusionProofs,
		},
		NewAsset:             btcOutputInfos[i].OutputAsset[0].Copy(), // TODO:
		RootOutputIndex:      uint32(outIndex),
		RootInternalKey:      btcOutputInfos[i].AddrResult.PubKey,
		RootTaprootAssetTree: btcOutputInfos[i].AddrResult.TapCommitment,
	}
}

func makeExclusionProofs(curID int, btcOutputInfos []*onchain.BtcOutputInfo) ([]*proof.TaprootProof, error) {
	curAsset := btcOutputInfos[curID].GetOutputAsset()[0].Copy()

	exclusionProofs := make([]*proof.TaprootProof, 0)
	for idx, exclusion := range btcOutputInfos {
		if idx == curID {
			continue
		}

		_, commitmentProof, err := exclusion.GetAddrResult().TapCommitment.CreateProof(
			curAsset.TapCommitmentKey(),
			curAsset.AssetCommitmentKey(),
		)
		if err != nil {
			return nil, err
		}

		exclusionProofs = append(exclusionProofs, &proof.TaprootProof{
			OutputIndex:     uint32(idx),
			InternalKey:     exclusion.GetAddrResult().PubKey,
			CommitmentProof: commitmentProof,
		})
	}

	return exclusionProofs, nil
}

func (t *Taproot) createReturnAsset(assetGenOutpoint *wire.OutPoint,
	assetName string,
	assetUTXOs *utxoasset.UnspentAssetResp, transferAsset []*asset.Asset) (*asset.Asset, error) {

	if len(assetUTXOs.UnspentOutpoints) == 0 || len(transferAsset) == 0 {
		return nil, errors.New("createReturnAsset: assetUTXOs or transferAsset is empty")
	}

	totalAmount := int32(0)
	transferAmount := int32(0)

	for _, a := range assetUTXOs.UnspentOutpoints {
		totalAmount += a.Amount
	}

	for _, a := range transferAsset {
		transferAmount += a.Amount
	}

	if (totalAmount - transferAmount) < 0 {
		return nil, errors.New("createReturnAsset: totalAmount - transferAmount < 0")
	}

	returnAsset := asset.New(*assetGenOutpoint, assetName,
		DEFAULT_RETURN_OUTPUT_INDEX, totalAmount-transferAmount,
		asset.ToSerialized(t.wif.PrivKey.PubKey()), nil,
	)

	return returnAsset, nil
}

func prepareAssets(assetGenOutpoint *wire.OutPoint,
	assetName string, amount []int32,
	receiverPubKey []asset.SerializedKey,
) []*asset.Asset {
	transferAsset := asset.New(*assetGenOutpoint, assetName,
		DEFAULT_TRANSFER_OUTPUT_INDEX, amount[0], receiverPubKey[0],
		nil,
	)

	return []*asset.Asset{transferAsset}
}

func (t *Taproot) prepareBtcOutputs(
	ctx context.Context,
	assetUTXOs *utxoasset.UnspentAssetResp,
	transferAsset []*asset.Asset,
	returnAsset *asset.Asset,
) ([]*onchain.BtcOutputInfo, *commitment.SplitCommitment, error) {
	var (
		returnPubKey   = asset.ToSerialized(t.wif.PrivKey.PubKey())
		btcOutputInfos = make([]*onchain.BtcOutputInfo, 0)
	)

	splitCommitment, err := createSplitCommitment(ctx, assetUTXOs, returnAsset, transferAsset)
	if err != nil {
		log.Println("[prepareBtcOutputs] createSplitCommitment(ctx, unspentAssets, returnAsset, transferAsset), err ", err)
		return nil, nil, err
	}

	returnCommitment, err := commitment.NewAssetCommitment(ctx, splitCommitment.RootAsset)
	if err != nil {
		log.Println("[prepareBtcOutputs] commitment.NewAssetCommitment(splitCommitment.RootAsset), err ", err)
		return nil, nil, err
	}

	tapReturnCommitment, err := commitment.NewTapCommitment(returnCommitment)
	returnOutputInfo, err := t.addressMaker.CreateTapAddr(returnPubKey, tapReturnCommitment)
	if err != nil {
		return nil, nil, err
	}
	btcOutputInfos = append(btcOutputInfos, onchain.NewBtcOutputInfo(returnOutputInfo, DEFAULT_OUTPUT_AMOUNT, returnAsset))

	rootLocator := commitment.NewLocatorByAsset(returnAsset)

	for locator, splitAsset := range splitCommitment.SplitAssets {
		if locator == *rootLocator {
			continue
		}

		splitAsset.Asset.PrevWitnesses[0].SplitCommitment = nil

		transferCommitment, err := commitment.NewAssetCommitment(ctx, &splitAsset.Asset)
		if err != nil {
			return nil, nil, err
		}

		tapTransferCommitment, err := commitment.NewTapCommitment(transferCommitment)
		if err != nil {
			return nil, nil, err
		}

		transferOutputInfo, err := t.addressMaker.CreateTapAddr(splitAsset.Asset.ScriptPubkey, tapTransferCommitment)
		if err != nil {
			return nil, nil, err
		}

		btcOutputInfos = append(btcOutputInfos, onchain.NewBtcOutputInfo(transferOutputInfo, DEFAULT_OUTPUT_AMOUNT, &splitAsset.Asset))
	}

	return btcOutputInfos, splitCommitment, nil
}

func createSplitCommitment(ctx context.Context,
	assetUTXOs *utxoasset.UnspentAssetResp,
	returnAsset *asset.Asset, transferAsset []*asset.Asset,
) (*commitment.SplitCommitment, error) {
	splitCommitmentInput, err := creatSplitCommitmentInputs(assetUTXOs)
	if err != nil {
		return nil, err
	}

	rootLocator := commitment.NewLocatorByAsset(returnAsset)

	externalLocators := make([]*commitment.SplitLocator, len(transferAsset))
	for idx, a := range transferAsset {
		externalLocators[idx] = commitment.NewLocatorByAsset(a)

	}

	return commitment.NewSplitCommitment(ctx, splitCommitmentInput, rootLocator, externalLocators...)
}

func creatSplitCommitmentInputs(
	assetUTXOs *utxoasset.UnspentAssetResp,
) ([]commitment.SplitCommitmentInput, error) {
	res := make([]commitment.SplitCommitmentInput, 0)

	assetGenOutpoint, err := wire.NewOutPointFromString(assetUTXOs.GenesisPoint.PrevOut)
	if err != nil {
		return nil, err
	}

	assetName := assetUTXOs.GenesisAsset.AssetName

	for _, input := range assetUTXOs.UnspentOutpoints {
		inputOutPoint, err := wire.NewOutPointFromString(input.Outpoint)
		if err != nil {
			return nil, err
		}

		curAsset := asset.New(*assetGenOutpoint,
			assetName, inputOutPoint.Index, input.Amount,
			asset.SerializedKey(input.ScriptKey),
			nil,
		)

		if len(input.SplitCommitmentRootHash) != 0 {
			curAsset.SplitCommitmentRoot = mssmt.NewComputedNode(
				mssmt.NodeHash(input.SplitCommitmentRootHash),
				uint64(input.SplitCommitmentRootValue),
			)
		}

		res = append(res, commitment.SplitCommitmentInput{
			Asset:    curAsset,
			OutPoint: *inputOutPoint,
		})
	}

	return res, nil
}
