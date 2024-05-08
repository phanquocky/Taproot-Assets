package commitment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"golang.org/x/exp/maps"
)

type CommittedAssets map[[32]byte]*asset.Asset

type AssetCommitment struct {
	TapKey [32]byte
	Root   *mssmt.BranchNode
	tree   mssmt.Tree
	Assets CommittedAssets
}

func NewAssetCommitment(ctx context.Context, assets ...*asset.Asset) (*AssetCommitment, error) {
	tree := mssmt.NewCompactedTree(mssmt.NewDefaultStore())
	committedAssets := make(CommittedAssets, len(assets))

	tapKey := assets[0].TapCommitmentKey()

	for _, a := range assets {
		if a.TapCommitmentKey() != tapKey {
			return nil, fmt.Errorf("asset ID mismatch: %x vs %x", a.TapCommitmentKey(), tapKey)
		}

		leaf, err := a.Leaf()
		if err != nil {
			return nil, err
		}

		key := a.AssetCommitmentKey()
		_, err = tree.Insert(ctx, key, leaf)
		if err != nil {
			return nil, fmt.Errorf("error inserting asset to assetcommitment tree: %w", err)
		}

		committedAssets[a.AssetCommitmentKey()] = a
	}

	root, err := tree.Root(ctx)
	if err != nil {
		return nil, err
	}

	return &AssetCommitment{
		TapKey: tapKey,
		Root:   root,
		tree:   tree,
		Assets: committedAssets,
	}, nil
}

func (c *AssetCommitment) TapCommitmentKey() [32]byte {
	return c.TapKey
}

func (c *AssetCommitment) Merge(other *AssetCommitment) error {
	if other.Assets == nil {
		return fmt.Errorf("cannot merge commitments without Assets")
	}
	if len(other.Assets) == 0 {
		return nil
	}

	for _, otherCommitment := range other.Assets {
		if err := c.Upsert(otherCommitment.Copy()); err != nil {
			return fmt.Errorf("error upserting other commitment: "+
				"%w", err)
		}
	}

	return nil
}

func (c *AssetCommitment) Upsert(newAsset *asset.Asset) error {
	if newAsset == nil {
		return errors.New("ErrNoAssets")
	}
	if newAsset.TapCommitmentKey() != c.TapKey {
		return errors.New("ErrAssetGenesisMismatch")
	}

	key := newAsset.AssetCommitmentKey()
	ctx := context.TODO()

	leaf, err := newAsset.Leaf()
	if err != nil {
		return err
	}

	_, err = c.tree.Insert(ctx, key, leaf)
	if err != nil {
		return err
	}

	c.Root, err = c.tree.Root(ctx)
	if err != nil {
		return err
	}

	c.Assets[key] = newAsset

	return nil
}

func (c *AssetCommitment) GetRoot() [sha256.Size]byte {
	left := c.Root.Left.NodeHash()
	right := c.Root.Right.NodeHash()

	h := sha256.New()
	_, _ = h.Write(c.TapKey[:])
	_, _ = h.Write(left[:])
	_, _ = h.Write(right[:])
	_ = binary.Write(h, binary.BigEndian, c.Root.NodeSum())
	return *(*[sha256.Size]byte)(h.Sum(nil))
}

func (c *AssetCommitment) TapCommitmentLeaf() *mssmt.LeafNode {
	root := c.GetRoot()
	sum := c.Root.NodeSum()

	var leaf bytes.Buffer
	_, _ = leaf.Write(root[:])
	_ = binary.Write(&leaf, binary.BigEndian, sum)
	return mssmt.NewLeafNode(leaf.Bytes(), sum)
}

func (c *AssetCommitment) GetAssets() CommittedAssets {
	assets := make(CommittedAssets, len(c.Assets))
	maps.Copy(assets, c.Assets)

	return assets
}

func (c *AssetCommitment) AssetProof(key [32]byte) (
	*asset.Asset, *mssmt.Proof, error) {

	if c.tree == nil {
		return nil, nil, fmt.Errorf("missing tree to compute proofs")
	}

	proof, err := c.tree.MerkleProof(context.TODO(), key)
	if err != nil {
		return nil, nil, err
	}

	return c.Assets[key], proof, nil
}

// TapLeaf constructs a new `TapLeaf` for this `TapCommitment`.
func (c *AssetCommitment) TapLeaf() txscript.TapLeaf {
	rootHash := c.Root.NodeHash()
	var rootSum [8]byte
	binary.BigEndian.PutUint64(rootSum[:], c.Root.NodeSum())
	leafParts := [][]byte{
		rootHash[:],
		rootSum[:],
	}
	leafScript := bytes.Join(leafParts, nil)
	return txscript.NewBaseTapLeaf(leafScript)
}
