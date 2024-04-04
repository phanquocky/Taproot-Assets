package utxo

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/quocky/taproot-asset/server/internal/domain/asset"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	"github.com/quocky/taproot-asset/server/internal/domain/genesis"
	utxoasset "github.com/quocky/taproot-asset/server/internal/domain/utxo_asset"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	utxoassetsdk "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	assetsdk "github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/utils"
	"golang.org/x/net/context"
)

type UseCase struct {
	assetRepo         asset.RepoInterface
	assetOutpointRepo assetoutpoint.RepoInterface
	genesisPointRepo  genesis.RepoInterface
}

func (u *UseCase) GetUnspentAssetsById(
	ctx context.Context,
	assetID string,
	amount int32,
	pubKey []byte,
) (*utxoassetsdk.UnspentAssetResp, error) {
	var (
		genesisAssets    []asset.GenesisAsset
		genesisAsset     asset.GenesisAsset
		unspentOutpoints []*assetoutpointmodel.UnspentOutpoint
		genesisPoint     genesis.GenesisPoint
		inputFilesBytes  [][]byte
		actualAmount     int32 = 0
	)

	assetIdBytes, err := hex.DecodeString(assetID)
	if err != nil {
		logger.Errorw("decode asset id fail", "asset_id", assetID, "err", err.Error())

		return nil, err
	}

	err = u.assetRepo.FindMany(ctx, map[string]any{"asset_id": assetIdBytes}, &genesisAssets)
	if err != nil {
		return nil, err
	}

	genesisAsset = genesisAssets[0]

	err = u.genesisPointRepo.FindOneByID(ctx, genesisAsset.GenesisPointID, &genesisPoint)
	if err != nil {
		return nil, err
	}

	allUnspentOutpoints, err := u.assetOutpointRepo.FindManyWithManagedUTXO(
		ctx,
		assetoutpoint.UnspentOutpointFilter{
			GenesisID: utils.ToPtr(genesisAsset.ID),
			Spent:     utils.ToPtr(false),
			ScriptKey: pubKey,
		},
	)
	if err != nil {
		return nil, err
	}

	unspentOutpoints, actualAmount = extractFromAllUnspentOutpoints(allUnspentOutpoints, amount)

	if actualAmount < amount {
		logger.Errorw("not enough amount", "actual_amount", actualAmount, "required_amount", amount)

		return nil, errors.New(fmt.Sprintf("not enough amount actual_amount %d required_amount %d", actualAmount, amount))
	}

	inputFilesBytes = make([][]byte, len(unspentOutpoints))
	for i, unspentOutpoint := range unspentOutpoints {
		filename := fmt.Sprintf(proof.LocatorFilePath, unspentOutpoint.ProofLocator)

		fileBytes, err := proof.FileBytesFromName(filename)
		if err != nil {
			logger.Errorw("get file bytes fail", "filename", filename, "err", err.Error())

			return nil, err
		}

		inputFilesBytes[i] = fileBytes
	}

	return &utxoassetsdk.UnspentAssetResp{
		GenesisAsset: assetsdk.GenesisAsset{
			AssetID:        genesisAsset.ID.String(),
			AssetName:      genesisAsset.AssetName,
			Supply:         genesisAsset.Supply,
			OutputIndex:    genesisAsset.OutputIndex,
			GenesisPointID: genesisAsset.GenesisPointID.String(),
		},
		UnspentOutpoints: unspentOutpoints,
		GenesisPoint: assetsdk.GenesisPoint{
			PrevOut:    genesisPoint.PrevOut,
			AnchorTxID: genesisPoint.AnchorTxID.String(),
		},
		InputFilesBytes: inputFilesBytes,
	}, nil
}

func extractFromAllUnspentOutpoints(allUnspentOutpoints []*assetoutpoint.UnspentOutpoint, amount int32) ([]*assetoutpointmodel.UnspentOutpoint, int32) {
	var (
		actualAmount     int32 = 0
		unspentOutpoints       = make([]*assetoutpointmodel.UnspentOutpoint, 0)
	)

	for _, uo := range allUnspentOutpoints {
		fmt.Println("unspentOutpoint", uo.TxID)

		unspentOutpoints = append(unspentOutpoints, &assetoutpointmodel.UnspentOutpoint{
			ID:                       uo.AssetOutpoint.ID.String(),
			GenesisID:                uo.GenesisID.String(),
			ScriptKey:                uo.ScriptKey,
			Amount:                   uo.Amount,
			SplitCommitmentRootHash:  uo.SplitCommitmentRootHash,
			SplitCommitmentRootValue: uo.SplitCommitmentRootValue,
			AnchorUtxoID:             uo.AnchorUtxoID.String(),
			ProofLocator:             uo.ProofLocator,
			Spent:                    uo.Spent,
			Outpoint:                 uo.Outpoint,
			AmtSats:                  uo.AmtSats,
			InternalKey:              uo.InternalKey,
			TaprootAssetRoot:         uo.TaprootAssetRoot,
			ScriptOutput:             uo.ScriptOutput,
			TxID:                     uo.TxID.String(),
		})

		actualAmount += uo.Amount
		if amount > 0 && actualAmount >= amount {
			break
		}
	}

	return unspentOutpoints, actualAmount
}

func NewUseCase(
	assetRepo asset.RepoInterface,
	assetOutpointRepo assetoutpoint.RepoInterface,
	genesisPointRepo genesis.RepoInterface,
) utxoasset.UseCaseInterface {
	return &UseCase{
		assetRepo:         assetRepo,
		assetOutpointRepo: assetOutpointRepo,
		genesisPointRepo:  genesisPointRepo,
	}
}
