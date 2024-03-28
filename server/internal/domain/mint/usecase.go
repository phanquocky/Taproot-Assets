package mint

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/quocky/taproot-asset/taproot/model/proof"
)

type UseCaseInterface interface {
	MintAsset(
		ctx context.Context,
		amountSats int32,
		tapScriptRootHash *chainhash.Hash,
		mintProof *proof.AssetProofs,
	) error
}
