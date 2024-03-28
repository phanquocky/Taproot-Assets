package mint

import (
	"bytes"
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/server/internal/domain/asset"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	chaintx "github.com/quocky/taproot-asset/server/internal/domain/chain_tx"
	"github.com/quocky/taproot-asset/server/internal/domain/common"
	"github.com/quocky/taproot-asset/server/internal/domain/genesis"
	manageutxo "github.com/quocky/taproot-asset/server/internal/domain/manage_utxo"
	"github.com/quocky/taproot-asset/server/internal/domain/mint"
	"github.com/quocky/taproot-asset/taproot/model/proof"
)

type UseCase struct {
	genesisPointRepo genesis.RepoInterface
	chainTxRepo      chaintx.RepoInterface
	assetRepo        asset.RepoInterface
	manageUtxoRepo   manageutxo.RepoInterface
	rpcClient        *rpcclient.Client
}

func (u *UseCase) MintAsset(ctx context.Context, amountSats int32, tapScriptRootHash *chainhash.Hash, mintProof *proof.AssetProofs) error {
	proofs := make([]proof.Proof, 0)
	for _, p := range *mintProof {
		_, err := p.Verify(ctx, nil)
		if err != nil {
			return err
		}

		proofs = append(proofs, *p)
	}

	file, err := proof.NewFile(proofs...)
	if err != nil {
		return err
	}

	locatorHash, err := file.Store()
	if err != nil {
		return err
	}

	for pubKey, p := range *mintProof {
		data := mint.InsertMintTxParams{
			Asset:             &p.Asset,
			OutputIdx:         int32(p.GenesisReveal.OutputIndex),
			AnchorTx:          &p.AnchorTx,
			AmountSats:        amountSats,
			AddressInfoPubkey: pubKey,
			TapScriptRootHash: tapScriptRootHash,
			ProofLocator:      locatorHash,
			MintProof:         p,
		}

		// insert to db
		_, err := u.insertTxMint(ctx, data)
		if err != nil {
			return err
		}

		// send raw tx
		_, err = u.rpcClient.SendRawTransaction(data.AnchorTx, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UseCase) insertTxMint(
	ctx context.Context,
	data mint.InsertMintTxParams,
) (*mint.InsertMintTxResult, error) {
	var (
		result  mint.InsertMintTxResult
		assetID = data.Asset.ID()
		txID    = data.AnchorTx.TxHash()
		txBytes bytes.Buffer

		dbTxs = make([]common.TransactionCallbackFunc, 0)
	)

	dbTxs = append(dbTxs,
		func(ctx context.Context) error {
			result.GenesisPoint = genesis.GenesisPoint{
				PrevOut:    "",
				AnchorTxID: 0,
			}

			docID, err := u.genesisPointRepo.InsertOne(ctx, result.GenesisPoint)

			result.GenesisPoint.ID = docID

			return err
		},
		func(ctx context.Context) error {
			result.AnchorTx = chaintx.ChainTx{
				TxID:     txID[:],
				AnchorTx: txBytes.Bytes(),
			}

			docID, err := u.chainTxRepo.InsertOne(ctx, result.AnchorTx)

			result.AnchorTx.ID = docID

			return err
		},
		func(ctx context.Context) error {
			result.GenesisAsset = asset.GenesisAsset{
				AssetID:        assetID[:],
				AssetName:      data.Asset.Name,
				Supply:         data.Asset.Amount,
				OutputIndex:    data.OutputIdx,
				GenesisPointID: 0,
			}

			docID, err := u.assetRepo.InsertOne(ctx, result.GenesisAsset)

			result.GenesisAsset.ID = docID

			return err
		},
		func(ctx context.Context) error {
			utxoOutpoint := wire.NewOutPoint(&txID, uint32(data.OutputIdx))
			result.ManagedUTXO = manageutxo.ManagedUtxo{
				Outpoint:         utxoOutpoint.String(),
				AmtSats:          data.AmountSats,
				InternalKey:      data.AddressInfoPubkey[:],
				TaprootAssetRoot: data.TapScriptRootHash[:],
				ScriptOutput:     data.AnchorTx.TxOut[data.OutputIdx].PkScript,
				TxID:             result.AnchorTx.ID,
			}

			docID, err := u.manageUtxoRepo.InsertOne(ctx, result.ManagedUTXO)

			result.ManagedUTXO.ID = docID

			return err
		},
		func(ctx context.Context) error {
			result.AssetOutpoint = assetoutpoint.AssetOutpoint{
				GenesisID:    result.GenesisAsset.ID,
				ScriptKey:    data.Asset.ScriptPubkey[:],
				Amount:       data.Asset.Amount,
				AnchorUtxoID: result.ManagedUTXO.ID,
				ProofLocator: data.ProofLocator[:],
			}

			docID, err := u.assetRepo.InsertOne(ctx, result.AssetOutpoint)

			result.AssetOutpoint.ID = docID

			return err
		},
	)

	if err := u.assetRepo.RunTransactions(ctx, dbTxs); err != nil {
		return nil, err
	}

	return &result, nil
}

func NewUseCase(
	assetRepo asset.RepoInterface,
	chainTxRepo chaintx.RepoInterface,
	genesisPointRepo genesis.RepoInterface,
	manageUtxoRepo manageutxo.RepoInterface,
	rpcClient *rpcclient.Client,
) mint.UseCaseInterface {
	return &UseCase{
		genesisPointRepo: genesisPointRepo,
		chainTxRepo:      chainTxRepo,
		assetRepo:        assetRepo,
		manageUtxoRepo:   manageUtxoRepo,
		rpcClient:        rpcClient,
	}
}
