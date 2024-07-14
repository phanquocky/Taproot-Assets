package commitment

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"

	"github.com/btcsuite/btcd/txscript"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"golang.org/x/exp/maps"
)

type AssetCommitments map[[32]byte]*AssetCommitment

// func (c AssetCommitments) String() string {
// 	return fmt.Sprintf("AssetCommitments{Key: %x, Value: %v}", maps.Keys(c), maps.Values(c))
// }

type TapCommitment struct {
	TreeRoot         *mssmt.BranchNode
	tree             mssmt.Tree
	AssetCommitments AssetCommitments
}

// func (c TapCommitment) String() string {
// 	return fmt.Sprintf("TapCommitment{TreeRoot: %s, AssetCommitments: %s}", c.TreeRoot, c.AssetCommitments)
// }

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
		AssetCommitments: assetCommitments,
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

func (c *TapCommitment) Assets() []*asset.Asset {
	var assets []*asset.Asset
	for _, commitment := range c.AssetCommitments {
		committedAssets := maps.Values(commitment.GetAssets())
		assets = append(assets, committedAssets...)
	}

	return assets
}

func (c *TapCommitment) CreateProof(tapCommitmentKey,
	assetCommitmentKey [32]byte) (*asset.Asset, *CommitmentProof, error) {

	if c.AssetCommitments == nil || c.tree == nil {
		return nil, nil, errors.New("missing asset commitments to compute proofs")
	}

	merkleProof, err := c.tree.MerkleProof(context.TODO(), tapCommitmentKey)
	if err != nil {
		return nil, nil, err
	}

	proof := &CommitmentProof{
		TapProof: &TapProof{
			Proof: *merkleProof,
		},
	}

	assetCommitment, ok := c.AssetCommitments[tapCommitmentKey]
	if !ok {
		return nil, proof, nil
	}

	a, assetProof, err := assetCommitment.AssetProof(assetCommitmentKey)
	if err != nil {
		return nil, nil, err
	}

	proof.AssetProof = &AssetProof{
		Proof:  *assetProof,
		TapKey: assetCommitment.TapKey,
	}

	return a, proof, nil
}
