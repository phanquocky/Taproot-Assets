package taproot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/http_model/transfer"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	assetoutpointmodel "github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
	"github.com/quocky/taproot-asset/taproot/utils"
)

func (t *Taproot) TransferAsset(receiverPubKey []asset.SerializedKey, assetId string, amount []int32) error {
	ctx := context.Background()

	err := t.verifyReceiverPubKey(receiverPubKey)
	if err != nil {
		return err
	}

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

	assetUTXOsOnchain, err := t.getAssetUTXOsOnchain(ctx, assetUTXOs)
	if err != nil {
		return err
	}

	assetGenOutpoint, err := wire.NewOutPointFromString(assetUTXOs.GenesisPoint.PrevOut)
	if err != nil {
		fmt.Println("wire.NewOutPointFromString(assetUTXOs.GenesisPoint.PrevOut) got error", err)

		return err
	}

	transferAssets := prepareAssets(assetGenOutpoint, &assetUTXOs.GenesisAsset, amount, receiverPubKey)

	returnAssets, err := t.createReturnAsset(assetGenOutpoint, &assetUTXOs.GenesisAsset, assetUTXOs, transferAssets)
	if err != nil {
		fmt.Println("t.createReturnAsset(assetGenOutpoint, assetUTXOs.GenesisAsset.AssetName, assetUTXOs, transferAssets) got error", err)

		return err
	}

	btcOutputInfos, _, err := t.prepareBtcOutputs(ctx, assetUTXOs, transferAssets, returnAssets.Assets)
	if err != nil {
		fmt.Println("t.createTransferAddresses(ctx, unspentAssets, transferAssets),  err ", err)
		return err
	}

	txIncludeOutPubKey, err := t.createTxOnChain(bestUTXOs, assetUTXOsOnchain,
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

	for _, btcOutputInfo := range btcOutputInfos {
		for _, a := range btcOutputInfo.GetOutputAsset() {
			if a.AssetID == nil || len(a.AssetID) != 32 {
				a.AssetID = assetUTXOs.GenesisAsset.AssetID
			}
		}
	}

	// fmt.Printf("anchortx: %s\n", txIncludeOutPubKey.Tx.TxHash().String())

	// fmt.Printf("anchortx bytes: %x\n", txIncludeOutPubKey.Tx.SerializeSize())

	data := transfer.TransferReq{
		GenesisAsset:     &assetUTXOs.GenesisAsset,
		AnchorTx:         txIncludeOutPubKey.Tx,
		AmtSats:          DEFAULT_OUTPUT_AMOUNT,
		BtcOutputInfos:   btcOutputInfos,
		UnspentOutpoints: assetUTXOs.UnspentOutpoints,
		Files:            files,
	}

	_, err = t.httpClient.R().SetBody(data).Post(os.Getenv("SERVER_BASE_URL") + "/transfer-asset")
	if err != nil {
		log.Println("t.httpClient.R().SetBody(data).Post(\"/transfer-asset\") got error", err)

		return err
	}

	return nil
}

func (t *Taproot) getAssetUTXOsOnchain(
	ctx context.Context,
	assetUTXOs *utxoasset.UnspentAssetResp,
) (*GetUnspentAssetsByIdResult, error) {
	assetUTXOsOnchain := make([]*OutPoint, 0)

	for _, u := range assetUTXOs.UnspentOutpoints {
		assetUTXOsOnchain = append(assetUTXOsOnchain, &OutPoint{
			Outpoint:         u.Outpoint,
			AmtSats:          int64(u.AmtSats),
			ScriptOutput:     u.ScriptOutput,
			InternalKey:      u.InternalKey,
			TaprootAssetRoot: u.TaprootAssetRoot,
		})
	}

	return &GetUnspentAssetsByIdResult{
		UnspentOutpoints: assetUTXOsOnchain,
	}, nil
}

func (t *Taproot) verifyReceiverPubKey(receiverPubKey []asset.SerializedKey) error {
	walletPubkey := t.GetPubKey()

	for _, key := range receiverPubKey {

		pubkey, err := key.ToPubKey()
		if err != nil {
			log.Println("key.ToPubKey() got error", err)
			return err
		}

		if walletPubkey == pubkey {
			return fmt.Errorf("verifyReceiverPubKey: receiverPubKey is the same as the sender's pubkey")
		}
	}

	return nil
}

func createFiles(
	inputFilesBytes [][]byte, // TODO: nen doi thanh map ?
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
			TapCommitment:   btcOutputInfos[i].AddrResult.GetTapCommitment(),
			ExclusionProofs: exclusionProofs,
		},
		NewAsset:             btcOutputInfos[i].OutputAsset[0].Copy(), // TODO:
		RootOutputIndex:      uint32(outIndex),
		RootInternalKey:      btcOutputInfos[outIndex].AddrResult.PubKey,
		RootTaprootAssetTree: btcOutputInfos[outIndex].AddrResult.GetTapCommitment(),
	}
}

func makeExclusionProofs(curID int, btcOutputInfos []*onchain.BtcOutputInfo) ([]*proof.TaprootProof, error) {
	curAsset := btcOutputInfos[curID].GetOutputAsset()[0].Copy()
	// curAsset.PrevWitnesses[0].SplitCommitment = nil

	exclusionProofs := make([]*proof.TaprootProof, 0)
	for idx, exclusion := range btcOutputInfos {
		if idx == curID {
			continue
		}

		_, commitmentProof, err := exclusion.GetAddrResult().GetTapCommitment().CreateProof(
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

type returnAssetsResp struct {
	Assets []*asset.Asset
}

func (t *Taproot) createReturnAsset(assetGenOutpoint *wire.OutPoint,
	genesisAsset *asset.GenesisAsset,
	assetUTXOs *utxoasset.UnspentAssetResp, transferAsset []*asset.Asset) (*returnAssetsResp, error) {

	if len(assetUTXOs.UnspentOutpoints) == 0 || len(transferAsset) == 0 {
		return nil, errors.New("createReturnAsset: assetUTXOs or transferAsset is empty")
	}

	passiveAssets, err := getPassiveAssets(assetUTXOs, transferAsset[0])
	if err != nil {
		return nil, err
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

	curAsset := asset.New(*assetGenOutpoint, genesisAsset.AssetName,
		DEFAULT_RETURN_OUTPUT_INDEX, totalAmount-transferAmount,
		asset.ToSerialized(t.wif.PrivKey.PubKey()), nil,
	)
	curAsset.AssetID = genesisAsset.AssetID
	returnAsset := []*asset.Asset{curAsset}

	returnAsset = append(returnAsset, passiveAssets...)

	return &returnAssetsResp{
		Assets: returnAsset,
	}, nil
}

func getPassiveAssets(utxOs *utxoasset.UnspentAssetResp, transferAsset *asset.Asset) ([]*asset.Asset, error) {
	activeAssetId := transferAsset.ID()
	passiveAssets := make([]*asset.Asset, 0)
	for _, u := range utxOs.UnspentOutpoints {
		for _, a := range u.RelatedAnchorAssets {
			var curAsset *assetoutpointmodel.AssetOutpoint
			err := json.Unmarshal([]byte(a), &curAsset)
			if err != nil {
				log.Println("json.Unmarshal([]byte(a), &curAsset) got error", err)
				return nil, err
			}

			if asset.ID(curAsset.Genesis.AssetID) != activeAssetId {

				//log.Println("curAsset.Genesis.AssetID: ", curAsset.Genesis.AssetID)
				passiveAssets = append(passiveAssets, &asset.Asset{
					AssetID:             curAsset.Genesis.AssetID,
					Amount:              curAsset.Amount,
					ScriptPubkey:        asset.SerializedKey(curAsset.ScriptKey),
					SplitCommitmentRoot: mssmt.NewComputedNode(mssmt.NodeHash(curAsset.SplitCommitmentRootHash), uint64(curAsset.SplitCommitmentRootValue)),
				})
			}

		}
	}

	return passiveAssets, nil
}

func prepareAssets(assetGenOutpoint *wire.OutPoint,
	genesisAsset *asset.GenesisAsset, amount []int32,
	receiverPubKey []asset.SerializedKey,
) []*asset.Asset {

	transferAsset := make([]*asset.Asset, len(amount))

	for idx, a := range amount { // TODO: SAI tu72 ngay cho64 nay2
		curAssset := asset.New(*assetGenOutpoint, genesisAsset.AssetName,
			DEFAULT_TRANSFER_OUTPUT_INDEX, a, receiverPubKey[idx], nil,
		)
		curAssset.AssetID = genesisAsset.AssetID

		transferAsset[idx] = curAssset
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

	rootLocator := commitment.NewLocatorByAsset(returnAsset[0])
	returnAsset[0] = splitCommitment.RootAsset

	ca := classifyAsset(returnAsset)

	returnAssetCommitments := createReturnAssetCommitments(ctx, ca)
	tapReturnCommitment, err := commitment.NewTapCommitment(returnAssetCommitments...)
	if err != nil {
		return nil, nil, err
	}

	returnOutputInfo, err := t.addressMaker.CreateTapAddr(returnPubKey, tapReturnCommitment)
	if err != nil {
		return nil, nil, err
	}
	btcOutputInfos = append(btcOutputInfos, onchain.NewBtcOutputInfo(returnOutputInfo, DEFAULT_OUTPUT_AMOUNT, returnAsset...))

	for locator, splitAsset := range splitCommitment.SplitAssets {
		if reflect.DeepEqual(locator, *rootLocator) {
			continue
		}

		splitAssetCopy := splitAsset.Asset.Copy()
		splitAssetCopy.PrevWitnesses[0].SplitCommitment = nil

		transferCommitment, err := commitment.NewAssetCommitment(ctx, splitAssetCopy)
		if err != nil {
			return nil, nil, err
		}

		tapTransferCommitment, err := commitment.NewTapCommitment(transferCommitment)
		if err != nil {
			return nil, nil, err
		}

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
		if len(assets) == 0 {
			continue
		}

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
