package proof

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
)

type Proof struct {
	PrevOut          wire.OutPoint
	AnchorTx         wire.MsgTx
	Asset            asset.Asset
	InclusionProof   TaprootProof
	ExclusionProofs  []*TaprootProof
	SplitRootProof   *TaprootProof
	AdditionalInputs []File
	GenesisReveal    *asset.Genesis
}
