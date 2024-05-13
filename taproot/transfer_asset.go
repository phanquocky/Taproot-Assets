package taproot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/http_model/transfer"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
	"github.com/quocky/taproot-asset/taproot/utils"
)

func (t *Taproot) TransferAsset(receiverPubKey []asset.SerializedKey, assetId string, amount []int32) error {
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

	totalAmount := utils.CalcSum(amount)

	assetUTXOs, err := t.GetAssetUTXOs(ctx, assetId, totalAmount)
	if err != nil {
		return err
	}

	assetGenOutpoint, err := wire.NewOutPointFromString(assetUTXOs.GenesisPoint.PrevOut)
	if err != nil {
		fmt.Println("wire.NewOutPointFromString(assetUTXOs.GenesisPoint.PrevOut) got error", err)

		return err
	}

	transferAssets := prepareAssets(assetGenOutpoint, assetUTXOs.GenesisAsset.AssetName, amount, receiverPubKey)

	returnAssets, err := t.createReturnAsset(assetGenOutpoint, assetUTXOs.GenesisAsset.AssetName, assetUTXOs, transferAssets)
	if err != nil {
		fmt.Println("t.createReturnAsset(assetGenOutpoint, assetUTXOs.GenesisAsset.AssetName, assetUTXOs, transferAssets) got error", err)

		return err
	}

	btcOutputInfos, _, err := t.prepareBtcOutputs(ctx, assetUTXOs, transferAssets, returnAssets.assets)
	if err != nil {
		fmt.Println("t.createTransferAddresses(ctx, unspentAssets, transferAssets),  err ", err)
		return err
	}

	txIncludeOutPubKey, err := t.createTxOnChain(bestUTXOs, nil,
		btcOutputInfos, btcutil.Amount(DEFAULT_FEE), true)
	if err != nil {
		return err
	}

	files, err := createFiles(
		assetUTXOs.InputFilesBytes,
		btcOutputInfos,
		txIncludeOutPubKey.Tx,
	)
	if err != nil {
		return err
	}

	fmt.Println("files: ", files)

	data := transfer.TransferReq{
		GenesisAsset:     &assetUTXOs.GenesisAsset,
		AnchorTx:         txIncludeOutPubKey.Tx,
		AmtSats:          DEFAULT_OUTPUT_AMOUNT,
		BtcOutputInfos:   btcOutputInfos,
		UnspentOutpoints: assetUTXOs.UnspentOutpoints,
		Files:            files,
	}

	postResp, err := t.httpClient.R().SetBody(data).Post(os.Getenv("SERVER_BASE_URL") + "/transfer-asset")
	if err != nil {
		log.Println("t.httpClient.R().SetBody(data).Post(\"/transfer-asset\") got error", err)

		return err
	}

	log.Println("[Transfer Asset] Post transfer asset success!", postResp)

	return nil
}

func createFiles(
	inputFilesBytes [][]byte, // TODO: nen doi thanh map ?
	btcOutputInfos []*onchain.BtcOutputInfo,
	tx *wire.MsgTx,
) ([]*proof.File, error) {
	curFiles := make([]*proof.File, len(btcOutputInfos))

	for i := range btcOutputInfos {
		log.Println("btcOutputInfos[i].GetOutputAsset()[0].Amount", btcOutputInfos[i].GetOutputAsset()[0].Amount)

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
	fmt.Println("btcOutputInfos[i].AddrResult", btcOutputInfos[i].AddrResult.TapCommitment)

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
		RootInternalKey:      btcOutputInfos[outIndex].AddrResult.PubKey,
		RootTaprootAssetTree: btcOutputInfos[outIndex].AddrResult.TapCommitment,
	}
}

func makeExclusionProofs(curID int, btcOutputInfos []*onchain.BtcOutputInfo) ([]*proof.TaprootProof, error) {
	log.Println("makeExclusionProofs: ")
	utils.PrintStruct(btcOutputInfos[0].OutputAsset[0])
	utils.PrintStruct(btcOutputInfos[1].OutputAsset[0])

	log.Println("----=-==makeExclusionProofs: ")
	log.Println(btcOutputInfos[curID].GetOutputAsset()[0].Copy().PrevWitnesses[0].SplitCommitment)

	curAsset := btcOutputInfos[curID].GetOutputAsset()[0].Copy()
	//curAsset.PrevWitnesses[0].SplitCommitment = nil

	fmt.Println("curAssetcurAssetcurAssetcurAssetcurAssetcurAssetcurAsset")
	utils.PrintStruct(curAsset)

	exclusionProofs := make([]*proof.TaprootProof, 0)
	for idx, exclusion := range btcOutputInfos {
		if idx == curID {
			continue
		}

		utils.PrintStruct(exclusion.OutputAsset[0])

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

type returnAssets struct {
	assets []*asset.Asset
}

func (t *Taproot) createReturnAsset(assetGenOutpoint *wire.OutPoint,
	assetName string,
	assetUTXOs *utxoasset.UnspentAssetResp, transferAsset []*asset.Asset) (*returnAssets, error) {

	if len(assetUTXOs.UnspentOutpoints) == 0 || len(transferAsset) == 0 {
		return nil, errors.New("createReturnAsset: assetUTXOs or transferAsset is empty")
	}

	passiveAssets, err := getPassiveAssets(assetUTXOs, transferAsset[0])
	if err != nil {

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

	returnAsset := []*asset.Asset{asset.New(*assetGenOutpoint, assetName,
		DEFAULT_RETURN_OUTPUT_INDEX, totalAmount-transferAmount,
		asset.ToSerialized(t.wif.PrivKey.PubKey()), nil,
	)}
	returnAsset = append(returnAsset, passiveAssets...)

	return &returnAssets{
		assets: passiveAssets,
	}, nil
}

func getPassiveAssets(utxOs *utxoasset.UnspentAssetResp, transferAsset *asset.Asset) ([]*asset.Asset, error) {
	activeAssetId := transferAsset.ID()
	passiveAssets := make([]*asset.Asset, 0)
	for _, u := range utxOs.UnspentOutpoints {
		for _, a := range u.RelatedAnchorAssets {
			var curAsset *asset.Asset

			err := json.Unmarshal([]byte(a), &curAsset)
			if err != nil {
				log.Println("json.Unmarshal([]byte(a), &curAsset) got error", err)
				return nil, err
			}
			if curAsset.ID() != activeAssetId {
				passiveAssets = append(passiveAssets, curAsset)
			}

		}
	}

	return passiveAssets, nil
}

func prepareAssets(assetGenOutpoint *wire.OutPoint,
	assetName string, amount []int32,
	receiverPubKey []asset.SerializedKey,
) []*asset.Asset {

	transferAsset := make([]*asset.Asset, len(amount))

	for idx, a := range amount {
		transferAsset[idx] = asset.New(*assetGenOutpoint, assetName,
			DEFAULT_TRANSFER_OUTPUT_INDEX, a, receiverPubKey[idx], nil,
		)
	}

	return transferAsset
}

func (t *Taproot) prepareBtcOutputs(
	ctx context.Context,
	assetUTXOs *utxoasset.UnspentAssetResp,
	transferAsset []*asset.Asset,
	returnAsset []*asset.Asset,
) ([]*onchain.BtcOutputInfo, *commitment.SplitCommitment, error) {
	var (
		returnPubKey   = asset.ToSerialized(t.wif.PrivKey.PubKey())
		btcOutputInfos = make([]*onchain.BtcOutputInfo, 0)
	)

	splitCommitment, err := createSplitCommitment(ctx, assetUTXOs, returnAsset[0], transferAsset) // returnAsset[0] is active asset
	if err != nil {
		log.Println("[prepareBtcOutputs] createSplitCommitment(ctx, unspentAssets, returnAsset, transferAsset), err ", err)
		return nil, nil, err
	}

	returnAsset[0] = splitCommitment.RootAsset

	ca := classifyAsset(returnAsset)

	returnAssetCommitments := createReturnAssetCommitments(ctx, ca)
	tapReturnCommitment, err := commitment.NewTapCommitment(returnAssetCommitments...)

	fmt.Println("tapReturnCommitment: ", tapReturnCommitment.TreeRoot.NodeHash(), tapReturnCommitment.TreeRoot.NodeSum())
	utils.PrintStruct(tapReturnCommitment)

	returnOutputInfo, err := t.addressMaker.CreateTapAddr(returnPubKey, tapReturnCommitment)
	if err != nil {
		return nil, nil, err
	}
	btcOutputInfos = append(btcOutputInfos, onchain.NewBtcOutputInfo(returnOutputInfo, DEFAULT_OUTPUT_AMOUNT, returnAsset...))

	rootLocator := commitment.NewLocatorByAsset(returnAsset[0])

	for locator, splitAsset := range splitCommitment.SplitAssets {
		if locator == *rootLocator {
			continue
		}

		log.Println("bdefore Copy: splitAsset.Asset.PrevWitnesses[0].SplitCommitment", splitAsset.Asset.PrevWitnesses[0].SplitCommitment)
		splitAssetCopy := splitAsset.Asset.Copy()
		splitAssetCopy.PrevWitnesses[0].SplitCommitment = nil

		log.Println("after Copy: splitAsset.Asset.PrevWitnesses[0].SplitCommitment", splitAsset.Asset.PrevWitnesses[0].SplitCommitment)

		transferCommitment, err := commitment.NewAssetCommitment(ctx, splitAssetCopy)
		if err != nil {
			return nil, nil, err
		}

		tapTransferCommitment, err := commitment.NewTapCommitment(transferCommitment)
		if err != nil {
			return nil, nil, err
		}

		fmt.Println("tapTransferCommitment: ", tapTransferCommitment.TreeRoot.NodeHash(), tapTransferCommitment.TreeRoot.NodeSum())
		utils.PrintStruct(tapTransferCommitment)

		transferOutputInfo, err := t.addressMaker.CreateTapAddr(splitAssetCopy.ScriptPubkey, tapTransferCommitment)
		if err != nil {
			return nil, nil, err
		}

		btcOutputInfos = append(btcOutputInfos, onchain.NewBtcOutputInfo(transferOutputInfo, DEFAULT_OUTPUT_AMOUNT, &splitAsset.Asset))
	}

	return btcOutputInfos, splitCommitment, nil
}

func createReturnAssetCommitments(ctx context.Context, ca map[[32]byte][]*asset.Asset) []*commitment.AssetCommitment {

	returnAssetCommitments := make([]*commitment.AssetCommitment, 0)
	for _, assets := range ca {
		assetCommitments, err := commitment.NewAssetCommitment(ctx, assets...)
		if err != nil {
			log.Println("[createReturnAssetCommitments] commitment.NewAssetCommitment(ctx, assets...), err ", err)
			return nil
		}

		returnAssetCommitments = append(returnAssetCommitments, assetCommitments)
	}

	return returnAssetCommitments
}

func classifyAsset(returnAsset []*asset.Asset) map[[32]byte][]*asset.Asset {
	ca := make(map[[32]byte][]*asset.Asset)
	for _, a := range returnAsset {
		if ca[a.AssetCommitmentKey()] == nil {
			ca[a.AssetCommitmentKey()] = make([]*asset.Asset, 0)
		}
		ca[a.TapCommitmentKey()] = append(ca[a.TapCommitmentKey()], a)
	}

	return ca
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
