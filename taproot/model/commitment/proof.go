package commitment

import (
	"errors"
	"github.com/lightninglabs/taproot-assets/fn"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"log"
)

var (
	// ErrInvalidTapscriptProof is an error returned upon attempting to
	// prove a malformed TapscriptProof.
	ErrInvalidTapscriptProof = errors.New("invalid tapscript proof")

	// ErrInvalidTaprootProof is an error returned upon verifying an invalid
	// Taproot proof.
	ErrInvalidTaprootProof = errors.New("invalid taproot proof")
)

type AssetProof struct {
	mssmt.Proof
	TapKey [32]byte
}

type TapProof struct {
	mssmt.Proof
}

type CommitmentProof struct {
	AssetProof *AssetProof
	TapProof   *TapProof
}

// DeriveByAssetInclusion derives the Asset commitment containing the
// provided asset. This consists of proving that an asset exists within the
// inner MS-SMT with the AssetProof, also known as the AssetCommitment.
func (p CommitmentProof) DeriveByAssetInclusion(asset *asset.Asset) (*TapCommitment,
	error) {

	if p.AssetProof == nil {
		return nil, errors.New("missing commitment proof")
	}

	// Use the asset proof to arrive at the asset commitment included within
	// the Taproot Asset commitment.
	assetCommitmentLeaf, err := asset.Leaf()
	if err != nil {
		return nil, err
	}

	assetProofRoot := p.AssetProof.Root(
		asset.AssetCommitmentKey(), assetCommitmentLeaf,
	)

	assetCommitment := &AssetCommitment{
		TapKey: p.AssetProof.TapKey,
		Root:   assetProofRoot,
	}

	// Use the Taproot Asset commitment proof to arrive at the Taproot Asset
	// commitment.
	tapProofRoot := p.TapProof.Root(
		assetCommitment.TapCommitmentKey(),
		assetCommitment.TapCommitmentLeaf(),
	)

	log.Println("tapProofRoot: ", tapProofRoot.NodeSum())
	log.Println("tapProofRoot: ", tapProofRoot.NodeHash())

	log.Printf("Derived asset inclusion proof for asset_id=%v, "+
		"asset_commitment_key=%x, asset_commitment_leaf=%s",
		asset.ID(), fn.ByteSlice(asset.AssetCommitmentKey()),
		assetCommitmentLeaf.NodeHash())

	return &TapCommitment{
		TreeRoot:         tapProofRoot,
		tree:             nil,
		assetCommitments: nil,
	}, nil
}

// DeriveByAssetExclusion derives the Taproot Asset commitment excluding the
// given asset identified by its key within an AssetCommitment. This consists of
// proving with the AssetProof that an asset does not exist within the inner
// MS-SMT, also known as the AssetCommitment. With the AssetCommitment obtained,
// the TapProof is used to prove that the AssetCommitment exists within
// the outer MS-SMT, also known as the TapCommitment.
func (p CommitmentProof) DeriveByAssetExclusion(assetCommitmentKey [32]byte) (
	*TapCommitment, error) {

	// Use the asset proof to arrive at the asset commitment included within
	// the Taproot Asset commitment.
	assetCommitmentLeaf := mssmt.EmptyLeafNode
	assetProofRoot := p.AssetProof.Root(
		assetCommitmentKey, assetCommitmentLeaf,
	)
	assetCommitment := &AssetCommitment{
		TapKey: p.AssetProof.TapKey,
		Root:   assetProofRoot,
	}

	// Use the Taproot Asset commitment proof to arrive at the Taproot Asset
	// commitment.
	tapProofRoot := p.TapProof.Root(
		assetCommitment.TapCommitmentKey(),
		assetCommitment.TapCommitmentLeaf(),
	)
	return &TapCommitment{
		TreeRoot:         tapProofRoot,
		tree:             nil,
		assetCommitments: nil,
	}, nil
}

// DeriveByAssetCommitmentExclusion derives the Taproot Asset commitment
// excluding the given asset commitment identified by its key within a
// TapCommitment. This consists of proving with the TapProof that an
// AssetCommitment does not exist within the outer MS-SMT, also known as the
// TapCommitment.
func (p CommitmentProof) DeriveByAssetCommitmentExclusion(tapCommitmentKey [32]byte) (
	*TapCommitment, error) {

	if p.AssetProof != nil {
		return nil, errors.New("attempting to prove an invalid asset " +
			"commitment exclusion")
	}

	// Use the Taproot Asset commitment proof to arrive at the Taproot Asset
	// commitment.
	tapProofRoot := p.TapProof.Root(
		tapCommitmentKey, mssmt.EmptyLeafNode,
	)
	return &TapCommitment{
		TreeRoot:         tapProofRoot,
		tree:             nil,
		assetCommitments: nil,
	}, nil
}
