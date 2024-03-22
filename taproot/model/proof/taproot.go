package proof

import "github.com/quocky/taproot-asset/taproot/model/asset"

type TapscriptProof struct {
	Bip86 bool
}

type TaprootProof struct {
	OutputIndex     uint32
	InternalKey     asset.SerializedKey
	CommitmentProof *commitment.CommitmentProof
	TapscriptProof  *TapscriptProof
}
