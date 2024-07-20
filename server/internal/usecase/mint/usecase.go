package mint

import (
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	chaintx "github.com/quocky/taproot-asset/server/internal/domain/chain_tx"
	"github.com/quocky/taproot-asset/server/internal/domain/common"
	"github.com/quocky/taproot-asset/server/internal/domain/genesis"
	genesisasset "github.com/quocky/taproot-asset/server/internal/domain/genesis_asset"
	manageutxo "github.com/quocky/taproot-asset/server/internal/domain/manage_utxo"
	"github.com/quocky/taproot-asset/server/internal/domain/mint"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	"github.com/quocky/taproot-asset/taproot/model/proof"
)

type UseCase struct {
	genesisPointRepo  genesis.RepoInterface
	chainTxRepo       chaintx.RepoInterface
	assetRepo         genesisasset.RepoInterface
	assetOutpointRepo assetoutpoint.RepoInterface
	manageUtxoRepo    manageutxo.RepoInterface
	rpcClient         *rpcclient.Client
}

func (u *UseCase) MintAsset(
	ctx context.Context,
	amountSats int32,
	tapScriptRootHash *chainhash.Hash,
	mintProof proof.AssetProofs,
) error {
	if len(mintProof) == 0 {
		return errors.New("mint proof empty")
	}

	locatorHash, err := generateLocatorHash(ctx, mintProof)
	if err != nil {
		logger.Errorw("create new locator hash fail", "err", err.Error())

		return err
	}

	chainTxID, genesisPointID, manageUtxoID, err := u.insertCommonComp(ctx, amountSats, tapScriptRootHash, mintProof[0])
	if err != nil {
		return err
	}

	for _, p := range mintProof {
		data := mint.InsertMintTxParams{
			Asset:             &p.Asset,
			OutputIdx:         int32(p.GenesisReveal.OutputIndex),
			AnchorTx:          &p.AnchorTx,
			AmountSats:        amountSats,
			AddressInfoPubkey: p.Asset.ScriptPubkey,
			TapScriptRootHash: tapScriptRootHash,
			ProofLocator:      locatorHash,
			MintProof:         p,
		}

		// insert to db
		_, err := u.insertDiffCompTxMint(ctx, data, *chainTxID, *genesisPointID, *manageUtxoID)
		if err != nil {
			return err
		}
	}

	log.Println("send raw tx 1")
	// send raw tx
	_, err = u.rpcClient.SendRawTransaction(&mintProof[0].AnchorTx, true)
	if err != nil {
		logger.Errorw("SendRawTransaction fail", err)

		return err
	}

	_, err = u.rpcClient.Generate(1)
	if err != nil {
		logger.Errorw("Generate fail", err)

		return err
	}

	return nil
}

func generateLocatorHash(ctx context.Context, mintProof proof.AssetProofs) ([32]byte, error) {
	proofs := make([]proof.Proof, 0)

	for _, p := range mintProof {
		_, err := p.Verify(ctx, nil)
		if err != nil {
			logger.Errorw("verify fail", "err", err.Error())

			return [32]byte{}, err
		}

		proofs = append(proofs, *p)
	}

	file, err := proof.NewFile(proofs...)
	if err != nil {
		logger.Errorw("create new file fail", "err", err.Error())

		return [32]byte{}, err
	}

	return file.Store()
}

func (u *UseCase) insertDiffCompTxMint(
	ctx context.Context,
	data mint.InsertMintTxParams,
	chainTxID, genesisPointID, manageUtxoID common.ID,
) (*mint.InsertMintTxResult, error) {
	var (
		result  = mint.InsertMintTxResult{}
		assetID = data.Asset.ID()

		dbTxs = make([]common.TransactionCallbackFunc, 0)
	)

	result.AnchorTx.ID = chainTxID
	result.GenesisPoint.ID = genesisPointID

	dbTxs = append(dbTxs,
		func(ctx context.Context) error {
			result.GenesisAsset = genesisasset.GenesisAsset{
				AssetID:        assetID[:],
				AssetName:      data.Asset.Name,
				Supply:         data.Asset.Amount,
				OutputIndex:    data.OutputIdx,
				GenesisPointID: result.GenesisPoint.ID,
			}

			docID, err := u.assetRepo.InsertOne(ctx, result.GenesisAsset)

			result.GenesisAsset.ID = docID

			return err
		},
		func(ctx context.Context) error {
			result.AssetOutpoint = assetoutpoint.AssetOutpoint{
				GenesisID:               result.GenesisAsset.ID,
				ScriptKey:               data.Asset.ScriptPubkey[:],
				Amount:                  data.Asset.Amount,
				AnchorUtxoID:            manageUtxoID,
				ProofLocator:            data.ProofLocator[:],
				SplitCommitmentRootHash: make([]byte, 32),
			}

			docID, err := u.assetOutpointRepo.InsertOne(ctx, result.AssetOutpoint)

			result.AssetOutpoint.ID = docID

			return err
		},
	)

	if err := u.assetRepo.RunTransactions(ctx, dbTxs); err != nil {
		return nil, err
	}

	return &result, nil
}

func (u *UseCase) insertCommonComp(
	ctx context.Context,
	amountSats int32,
	tapScriptRootHash *chainhash.Hash,
	p *proof.Proof,
) (*common.ID, *common.ID, *common.ID, error) {
	var (
		txHash  = p.AnchorTx.TxHash()
		txBytes bytes.Buffer
	)

	if err := p.AnchorTx.Serialize(&txBytes); err != nil {
		logger.Errorw("p.AnchorTx.Serialize fail", err.Error())

		return nil, nil, nil, err
	}

	chainTxID, err := u.chainTxRepo.InsertOne(ctx, chaintx.ChainTx{
		TxID:     txHash[:],
		AnchorTx: txBytes.Bytes(),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	genesisPointID, err := u.genesisPointRepo.InsertOne(ctx, genesis.GenesisPoint{
		PrevOut:    p.Asset.FirstPrevOut.String(),
		AnchorTxID: chainTxID,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	utxoOutpoint := wire.NewOutPoint(&txHash, p.Asset.OutputIndex)
	manageUtxoID, err := u.manageUtxoRepo.InsertOne(ctx, manageutxo.ManagedUtxo{
		Outpoint:         utxoOutpoint.String(),
		AmtSats:          amountSats,
		InternalKey:      p.Asset.ScriptPubkey[:],
		TaprootAssetRoot: tapScriptRootHash[:],
		ScriptOutput:     p.AnchorTx.TxOut[p.Asset.OutputIndex].PkScript,
		TxID:             chainTxID,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return &chainTxID, &genesisPointID, &manageUtxoID, nil
}

func NewUseCase(
	assetRepo genesisasset.RepoInterface,
	assetOutpointRepo assetoutpoint.RepoInterface,
	chainTxRepo chaintx.RepoInterface,
	genesisPointRepo genesis.RepoInterface,
	manageUtxoRepo manageutxo.RepoInterface,
	rpcClient *rpcclient.Client,
) mint.UseCaseInterface {
	return &UseCase{
		assetOutpointRepo: assetOutpointRepo,
		genesisPointRepo:  genesisPointRepo,
		chainTxRepo:       chainTxRepo,
		assetRepo:         assetRepo,
		manageUtxoRepo:    manageUtxoRepo,
		rpcClient:         rpcClient,
	}
}
