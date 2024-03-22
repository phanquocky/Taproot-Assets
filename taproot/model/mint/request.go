package mint

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type MintAssetReq struct {
	AmountSats        int32              `json:"amount_sats"`
	TapScriptRootHash *chainhash.Hash    `json:"tap_script_root_hash"`
	MintProof         *proof.AssetProofs `json:"mint_proof"`
}
