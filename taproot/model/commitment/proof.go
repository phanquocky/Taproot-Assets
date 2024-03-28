package commitment

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"

	"github.com/lightninglabs/taproot-assets/fn"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
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

type CommitmentProofByte struct {
	AssetProof []byte
	TapProof   []byte
	TapKey     [32]byte
}

// MarshalJSON function custom CommitmentProof when using json.Marshal function
func (u CommitmentProof) MarshalJSON() ([]byte, error) {
	var assetProofBytes bytes.Buffer

	err := u.AssetProof.Compress().Encode(&assetProofBytes)
	if err != nil {
		log.Println("[MarshalJSON] u.Proof.Compress().Encode(&proofBytes), err ", err)

		return nil, err
	}

	var tapProofBytes bytes.Buffer
	if err := u.TapProof.Compress().Encode(&tapProofBytes); err != nil {
		log.Println("[MarshalJSON] u.Proof.Compress().Encode(&proofBytes), err ", err)

		return nil, err
	}

	cbyte, _ := json.Marshal(CommitmentProofByte{
		AssetProof: assetProofBytes.Bytes(),
		TapProof:   tapProofBytes.Bytes(),
		TapKey:     u.AssetProof.TapKey,
	})

	log.Println("byte ", cbyte)

	return json.Marshal(CommitmentProofByte{
		AssetProof: assetProofBytes.Bytes(),
		TapProof:   tapProofBytes.Bytes(),
		TapKey:     u.AssetProof.TapKey,
	})
}

func (b *CommitmentProof) UnmarshalJSON(data []byte) error {
	var (
		commitBytes CommitmentProofByte
		assetProof  mssmt.CompressedProof
		tapProof    mssmt.CompressedProof
	)

	if err := json.Unmarshal(data, &commitBytes); err != nil {
		log.Println("err := json.Unmarshal(data, &commitBytes), err ", err)

		return err
	}

	//b.AssetProof.Compress()

	// commitBytes.Proof
	err := assetProof.Decode(bytes.NewReader(commitBytes.AssetProof))
	if err != nil {
		log.Println("assetProof.Decode(bytes.NewReader(commitBytes.AssetProof))", err.Error())

		return err
	}

	bufAssetProof, err := assetProof.Decompress()
	if err != nil {
		log.Println("err := compressProof.Decompress(), err ", err)

		return err
	}

	b.AssetProof = &AssetProof{
		Proof:  *bufAssetProof,
		TapKey: commitBytes.TapKey,
	}

	//b.TapProof.Compress()

	// commitBytes.Proof
	tapProof.Decode(bytes.NewReader(commitBytes.TapProof))
	bufTapAssetProof, err := tapProof.Decompress()
	if err != nil {
		log.Println("err := compressProof.Decompress(), err ", err)

		return err
	}
	b.TapProof = &TapProof{
		Proof: *bufTapAssetProof,
	}

	return nil
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
