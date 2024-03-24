package commitment

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"golang.org/x/exp/maps"
)

type AssetCommitments map[[32]byte]*AssetCommitment

type TapCommitment struct {
	TreeRoot         *mssmt.BranchNode
	tree             mssmt.Tree
	assetCommitments AssetCommitments
}

func NewTapCommitment(newCommitments ...*AssetCommitment) (*TapCommitment,
	error) {

	tree := mssmt.NewCompactedTree(mssmt.NewDefaultStore())
	assetCommitments := make(AssetCommitments, len(newCommitments))
	for idx := range newCommitments {
		assetCommitment := newCommitments[idx]
		key := assetCommitment.TapCommitmentKey()

		existingCommitment, ok := assetCommitments[key]
		if ok {
			err := existingCommitment.Merge(assetCommitment)
			if err != nil {
				return nil, err
			}

			assetCommitment = existingCommitment
			assetCommitments[key] = existingCommitment

		}

		leaf := assetCommitment.TapCommitmentLeaf()

		_, err := tree.Insert(context.TODO(), key, leaf)
		if err != nil {
			return nil, err
		}

		assetCommitments[key] = assetCommitment
	}

	root, err := tree.Root(context.Background())
	if err != nil {
		return nil, err
	}

	return &TapCommitment{
		TreeRoot:         root,
		assetCommitments: assetCommitments,
		tree:             tree,
	}, nil
}

func (c *TapCommitment) TapLeaf() txscript.TapLeaf {
	rootHash := c.TreeRoot.NodeHash()
	var rootSum [8]byte
	binary.BigEndian.PutUint64(rootSum[:], c.TreeRoot.NodeSum())
	leafParts := [][]byte{
		rootHash[:],
		rootSum[:],
	}
	leafScript := bytes.Join(leafParts, nil)
	return txscript.NewBaseTapLeaf(leafScript)
}

// CommittedAssets returns the set of assets committed to in the Taproot Asset
// commitment.
func (c *TapCommitment) CommittedAssets() []*asset.Asset {
	var assets []*asset.Asset
	for _, commitment := range c.assetCommitments {
		commitmentClone := commitment

		committedAssets := maps.Values(commitmentClone.Assets())
		assets = append(assets, committedAssets...)
	}

	return assets
}

// Proof computes the full TapCommitment merkle proof for the asset leaf
// located at `assetCommitmentKey` within the AssetCommitment located at
// `tapCommitmentKey`.
func (c *TapCommitment) Proof(tapCommitmentKey,
	assetCommitmentKey [32]byte) (*asset.Asset, *CommitmentProof, error) {

	if c.assetCommitments == nil || c.tree == nil {
		panic("missing asset commitments to compute proofs")
	}

	// TODO(bhandras): thread the context through.
	merkleProof, err := c.tree.MerkleProof(context.TODO(), tapCommitmentKey)
	if err != nil {
		return nil, nil, err
	}

	proof := &CommitmentProof{
		TaprootAssetProof: &TaprootAssetProof{
			Proof: *merkleProof,
		},
	}

	// If the corresponding AssetCommitment does not exist, return the Proof
	// as is.

	fmt.Println("key2: ", tapCommitmentKey)

	assetCommitment, ok := c.assetCommitments[tapCommitmentKey]
	if !ok {
		return nil, proof, nil
	}

	// Otherwise, compute the AssetProof and include it in the result. It's
	// possible for the asset to not be found, leading to a non-inclusion
	// proof.
	a, assetProof, err := assetCommitment.AssetProof(assetCommitmentKey)
	if err != nil {
		return nil, nil, err
	}

	proof.AssetProof = &AssetProof{
		Proof:  *assetProof,
		TapKey: assetCommitment.TapKey,
	}

	fmt.Println("aaaaaaaaaaaa: ", a)

	return a, proof, nil
}
