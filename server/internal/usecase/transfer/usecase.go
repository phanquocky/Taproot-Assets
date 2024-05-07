package transfer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	chaintx "github.com/quocky/taproot-asset/server/internal/domain/chain_tx"
	"github.com/quocky/taproot-asset/server/internal/domain/common"
	manageutxo "github.com/quocky/taproot-asset/server/internal/domain/manage_utxo"
	"github.com/quocky/taproot-asset/server/internal/domain/transfer"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	assetoutpointmodel "github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
	"github.com/quocky/taproot-asset/taproot/utils"
	"golang.org/x/net/context"
)

type UseCase struct {
	assetOutpointRepo assetoutpoint.RepoInterface
	chainTXRepo       chaintx.RepoInterface
	manageUtxoRepo    manageutxo.RepoInterface
	rpcClient         *rpcclient.Client
}

func (u *UseCase) TransferAsset(
	ctx context.Context,
	genesisAsset *asset.GenesisAsset,
	anchorTx *wire.MsgTx,
	amtSats int32,
	btcOutputInfos []*onchain.BtcOutputInfo,
	unspentOutpoints []*assetoutpointmodel.UnspentOutpoint,
	files []*proof.File,
) error {
	if err := u.insertDBTransferTx(
		ctx,
		genesisAsset,
		anchorTx,
		amtSats,
		btcOutputInfos,
		unspentOutpoints,
		files,
	); err != nil {
		return err
	}

	_, err := u.rpcClient.SendRawTransaction(anchorTx, true)
	if err != nil {
		logger.Errorw("rpcClient.SendRawTransaction fail", "tx_hash", anchorTx.TxHash(), "err", err)

		return err
	}

	for _, unspentOutpoint := range unspentOutpoints {
		filename := fmt.Sprintf(proof.LocatorFilePath, unspentOutpoint.ProofLocator)

		if err := os.Remove(filename); err != nil {
			logger.Errorw("remove proof file fail", "filename", filename, "err", err)

			return err
		}
	}

	return nil
}

func (u *UseCase) insertDBTransferTx(
	ctx context.Context,
	genesisAsset *asset.GenesisAsset,
	anchorTx *wire.MsgTx,
	amtSats int32,
	btcOutputInfos []*onchain.BtcOutputInfo,
	unspentOutpoints []*assetoutpointmodel.UnspentOutpoint,
	files []*proof.File,
) error {
	var (
		txBytes bytes.Buffer
		txID    = anchorTx.TxHash()
		err     error
	)

	if err := anchorTx.Serialize(&txBytes); err != nil {
		logger.Errorw("anchorTx.Serialize err", err)

		return err
	}

	chainTxID, err := u.chainTXRepo.InsertOne(ctx, &chaintx.ChainTx{
		TxID:     txID[:],
		AnchorTx: txBytes.Bytes(),
	})
	if err != nil {
		return err
	}

	for outID, btcOut := range btcOutputInfos {
		locatorName, err := files[outID].Store()
		if err != nil {
			logger.Errorw("locatorName, err := arg.Files[outIndex].Store()", "err", err)

			return err
		}

		utxoID, err := u.manageUtxoRepo.InsertOne(ctx, &manageutxo.ManagedUtxo{
			Outpoint:         wire.NewOutPoint(&txID, uint32(outID)).String(),
			AmtSats:          amtSats,
			InternalKey:      btcOut.GetAddrResult().PubKey[:],
			TaprootAssetRoot: btcOut.GetAddrResult().TapScriptRootHash[:],
			ScriptOutput:     anchorTx.TxOut[outID].PkScript,
			TxID:             chainTxID,
		})
		if err != nil {
			return err
		}

		btcOutAssets := btcOut.GetOutputAsset()
		// little confused
		curAsset := btcOutAssets[0]

		tapCommitmentBytes, err := json.Marshal(btcOut.AddrResult.TapCommitment)
		if err != nil {
			return errors.New("[insertDBTransferTx] marshal tap commitment fail " + err.Error())
		}

		insertAssetOutpointParam := assetoutpoint.AssetOutpoint{
			GenesisID:    common.ID(genesisAsset.GenesisPointID),
			ScriptKey:    curAsset.ScriptPubkey[:],
			Amount:       curAsset.Amount,
			AnchorUtxoID: utxoID,
			ProofLocator: locatorName[:],
			Spent:        false,
			// TODO duyba
			TapCommitment: tapCommitmentBytes,
		}

		if curAsset.SplitCommitmentRoot != nil {
			nodeHash := curAsset.SplitCommitmentRoot.NodeHash()

			insertAssetOutpointParam.SplitCommitmentRootValue = int32(curAsset.SplitCommitmentRoot.NodeSum())
			insertAssetOutpointParam.SplitCommitmentRootHash = nodeHash[:]
		}

		_, err = u.assetOutpointRepo.InsertOne(ctx, insertAssetOutpointParam)
		if err != nil {
			return err
		}
	}

	unspentIDs := make([]common.ID, len(unspentOutpoints))
	for i, uo := range unspentOutpoints {
		unspentIDs[i] = common.ID(uo.ID)
	}

	err = u.assetOutpointRepo.UpdateMany(
		ctx,
		assetoutpoint.UnspentOutpointFilter{
			IDs: &common.InOperator{Values: utils.ToSliceAny(unspentIDs)},
		},
		assetoutpoint.UnspentOutpointUpdate{
			Set: &assetoutpoint.UnspentOutpointSetUpdate{
				Spent: utils.ToPtr(true),
			},
		},
	)

	return nil
}

func NewUseCase(
	assetOutpointRepo assetoutpoint.RepoInterface,
	chainTXRepo chaintx.RepoInterface,
	manageUtxoRepo manageutxo.RepoInterface,
	rpcClient *rpcclient.Client,
) transfer.UseCaseInterface {
	return &UseCase{
		assetOutpointRepo: assetOutpointRepo,
		chainTXRepo:       chainTXRepo,
		manageUtxoRepo:    manageUtxoRepo,
		rpcClient:         rpcClient,
	}
}
