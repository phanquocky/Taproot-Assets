package utxo

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	"github.com/quocky/taproot-asset/server/internal/domain/genesis"
	genesisasset "github.com/quocky/taproot-asset/server/internal/domain/genesis_asset"
	utxoasset "github.com/quocky/taproot-asset/server/internal/domain/utxo_asset"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	utxoassetsdk "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	assetsdk "github.com/quocky/taproot-asset/taproot/model/asset"
	assetoutpointmodel "github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/utils"
	"golang.org/x/net/context"
)

type UseCase struct {
	genesisAssetRepo  genesisasset.RepoInterface
	assetOutpointRepo assetoutpoint.RepoInterface
	genesisPointRepo  genesis.RepoInterface
}

func (u *UseCase) ListAllAssetsWithAmount(
	ctx context.Context,
	pubkey []byte,
) (utxoassetsdk.ListAssetsResp, error) {
	return u.genesisAssetRepo.FindAvailableAssetsWithAmount(ctx, pubkey)
}

func (u *UseCase) GetUnspentAssetsById(
	ctx context.Context,
	assetID string,
	amount int32,
	pubKey []byte,
) (*utxoassetsdk.UnspentAssetResp, error) {
	var (
		genesisAsset     genesisasset.GenesisAsset
		unspentOutpoints []*assetoutpointmodel.UnspentOutpoint
		genesisPoint     genesis.GenesisPoint
		inputFilesBytes  [][]byte
		actualAmount     int32 = 0
	)

	assetIdBytes, err := hex.DecodeString(assetID)
	if err != nil {
		logger.Errorw("decode genesis_asset id fail", "asset_id", assetID, "err", err.Error())

		return nil, err
	}

	log.Printf("assetIdBytes %x", assetIdBytes)

	err = u.genesisAssetRepo.FindOne(ctx, map[string]any{"asset_id": assetIdBytes}, &genesisAsset)
	if err != nil {
		log.Println("find one fail", err)
		return nil, err
	}

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
		fmt.Println("unspentOutpoint", uo.TxID, uo.Amount)

		filename := fmt.Sprintf(proof.LocatorFilePath, uo.ProofLocator)

		fileBytes, err := proof.FileBytesFromName(filename)
		if err != nil {
			logger.Errorw("get file bytes fail", "filename", filename, "err", err.Error())

			return nil, 0
		}

		relatedAnchorAssets := make([][]byte, len(uo.RelatedAssets))
		relatedAnchorAssetProofs := make([][]byte, len(uo.RelatedAssets))

		for _, ra := range uo.RelatedAssets {
			raBytes, err := json.Marshal(ra)
			if err != nil {
				logger.Errorw("marshal related genesis_asset fail", "related_asset_id", ra.ID.String(), "err", err.Error())

				return nil, 0
			}

			relatedAnchorAssets = append(relatedAnchorAssets, raBytes)

			filenameRa := fmt.Sprintf(proof.LocatorFilePath, ra.ProofLocator)

			fileByteRas, err := proof.FileBytesFromName(filenameRa)
			if err != nil {
				logger.Errorw("get file bytes fail", "filename", filename, "err", err.Error())

				return nil, 0
			}

			relatedAnchorAssetProofs = append(relatedAnchorAssetProofs, fileByteRas)
		}

		unspentOutpoints = append(unspentOutpoints, &assetoutpointmodel.UnspentOutpoint{
			ID:                       uo.AssetOutpoint.ID.String(),
			GenesisID:                uo.GenesisID.String(),
			ScriptKey:                uo.ScriptKey,
			Amount:                   uo.Amount,
			SplitCommitmentRootHash:  uo.SplitCommitmentRootHash,
			SplitCommitmentRootValue: uo.SplitCommitmentRootValue,
			AnchorUtxoID:             uo.AnchorUtxoID.String(),
			ProofLocator:             uo.ProofLocator,
			Proof:                    fileBytes,
			Spent:                    uo.Spent,
			Outpoint:                 uo.Outpoint,
			AmtSats:                  uo.AmtSats,
			InternalKey:              uo.InternalKey,
			TaprootAssetRoot:         uo.TaprootAssetRoot,
			ScriptOutput:             uo.ScriptOutput,
			TxID:                     uo.TxID.String(),
			RelatedAnchorAssets:      relatedAnchorAssets,
			RelatedAnchorAssetProofs: relatedAnchorAssetProofs,
		})

		actualAmount += uo.Amount
		if amount > 0 && actualAmount >= amount {
			break
		}
	}

	return unspentOutpoints, actualAmount
}

func NewUseCase(
	genesisAssetRepo genesisasset.RepoInterface,
	assetOutpointRepo assetoutpoint.RepoInterface,
	genesisPointRepo genesis.RepoInterface,
) utxoasset.UseCaseInterface {
	return &UseCase{
		genesisAssetRepo:  genesisAssetRepo,
		assetOutpointRepo: assetOutpointRepo,
		genesisPointRepo:  genesisPointRepo,
	}
}
