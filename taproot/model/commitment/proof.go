package commitment

import "github.com/quocky/taproot-asset/taproot/model/mssmt"

type AssetProof struct {
	mssmt.Proof
	TapKey [32]byte
}

type TaprootAssetProof struct {
	mssmt.Proof
}

type CommitmentProof struct {
	AssetProof        *AssetProof
	TaprootAssetProof *TaprootAssetProof
}
