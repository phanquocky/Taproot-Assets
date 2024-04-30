package mint

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/proof"
)

type MintAssetReq struct {
	AmountSats        int32             `json:"amount_sats"`
	TapScriptRootHash *chainhash.Hash   `json:"tap_script_root_hash"`
	MintProof         proof.AssetProofs `json:"mint_proof"`
	// TODO: Add TapCommitment field
	TapCommitment *commitment.TapCommitment `json:"tap_commitment"`
}
