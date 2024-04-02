package commitment

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"log"

	"github.com/btcsuite/btcd/wire"
)

var (
	// ErrDuplicateSplitOutputIndex is an error returned when duplicate
	// split output indices are detected.
	ErrDuplicateSplitOutputIndex = errors.New(
		"found locator with duplicate output index",
	)

	// ErrInvalidSplitAmount is an error returned when a split amount is
	// invalid (e.g. splits do not fully consume input amount).
	ErrInvalidSplitAmount = errors.New("invalid split amounts")

	// ErrInvalidSplitLocator is returned if a new split is attempted to be
	// created w/o a valid external split locator.
	ErrInvalidSplitLocator = errors.New(
		"at least one locator should be specified",
	)

	// ErrInvalidSplitLocatorCount is returned if a collectible split is
	// attempted with a count of external split locators not equal to one.
	ErrInvalidSplitLocatorCount = errors.New(
		"exactly one locator should be specified",
	)

	// ErrInvalidScriptKey is an error returned when a root locator has zero
	// value but does not use the correct un-spendable script key.
	ErrInvalidScriptKey = errors.New(
		"invalid script key for zero-amount locator",
	)

	// ErrZeroSplitAmount is an error returned when a non-root split locator
	// has zero amount.
	ErrZeroSplitAmount = errors.New(
		"split locator has zero amount",
	)

	// ErrNonZeroSplitAmount is an error returned when a root locator uses
	// an un-spendable script key but has a non-zero amount.
	ErrNonZeroSplitAmount = errors.New(
		"un-spendable root locator has non-zero amount",
	)
)

type SplitLocator struct {
	OutputIndex uint32
	AssetID     asset.ID
	ScriptKey   asset.SerializedKey
	Amount      int32
}

func NewLocatorByAsset(a *asset.Asset) *SplitLocator {
	return &SplitLocator{
		OutputIndex: a.OutputIndex,
		AssetID:     a.ID(),
		ScriptKey:   a.ScriptPubkey,
		Amount:      a.Amount,
	}
}

func (l SplitLocator) Hash() [sha256.Size]byte {
	h := sha256.New()
	_ = binary.Write(h, binary.BigEndian, l.OutputIndex)
	_, _ = h.Write(l.AssetID[:])
	_, _ = h.Write(l.ScriptKey.SchnorrSerialized())
	return *(*[sha256.Size]byte)(h.Sum(nil))
}

type SplitAsset struct {
	asset.Asset
	OutputIndex uint32
}

type InputSet map[asset.PrevID]*asset.Asset
type SplitSet map[SplitLocator]*SplitAsset

type SplitCommitment struct {
	PrevAssets  InputSet
	RootAsset   *asset.Asset
	SplitAssets SplitSet
	Tree        mssmt.Tree
}

type SplitCommitmentInput struct {
	Asset    *asset.Asset
	OutPoint wire.OutPoint
}

func NewSplitCommitment(ctx context.Context, inputs []SplitCommitmentInput,
	rootLocator *SplitLocator,
	externalLocators ...*SplitLocator) (*SplitCommitment, error) {

	totalInputAmount := int32(0)
	for idx := range inputs {
		input := inputs[idx]
		totalInputAmount += input.Asset.Amount
	}

	if len(externalLocators) == 0 {
		return nil, ErrInvalidSplitLocator
	}

	locators := append(externalLocators, rootLocator)
	splitAssets := make(SplitSet, len(locators))

	splitTree := mssmt.NewCompactedTree(mssmt.NewDefaultStore())
	remainingAmount := totalInputAmount
	rootIdx := len(locators) - 1

	addAssetSplit := func(locator *SplitLocator) error {
		assetSplit := inputs[0].Asset.Copy()
		assetSplit.Amount = locator.Amount

		assetSplit.ScriptPubkey = locator.ScriptKey
		assetSplit.PrevWitnesses = []asset.Witness{{
			PrevID:          &asset.ZeroPrevID,
			SplitCommitment: nil,
		}}
		assetSplit.SplitCommitmentRoot = nil

		splitAssets[*locator] = &SplitAsset{
			Asset:       *assetSplit,
			OutputIndex: locator.OutputIndex,
		}

		splitKey := locator.Hash()
		splitLeaf, err := assetSplit.Leaf()
		if err != nil {
			return err
		}

		_, err = splitTree.Insert(ctx, splitKey, splitLeaf)
		if err != nil {
			return err
		}

		if remainingAmount < locator.Amount {
			return ErrInvalidSplitAmount
		}
		remainingAmount -= locator.Amount

		return nil
	}

	for idx := range locators {
		locator := locators[idx]
		if idx != rootIdx && locator.Amount == 0 {
			return nil, ErrZeroSplitAmount
		}

		if err := addAssetSplit(locator); err != nil {
			return nil, err
		}
	}
	if remainingAmount != 0 {
		return nil, ErrInvalidSplitAmount
	}

	var err error
	rootAsset := splitAssets[*rootLocator].Copy()

	inputSet := make(InputSet)
	rootAsset.PrevWitnesses = make([]asset.Witness, len(inputs))

	for idx := range inputs {
		input := inputs[idx]
		inAsset := input.Asset
		prevID := &asset.PrevID{
			OutPoint:  input.OutPoint,
			ID:        inAsset.Genesis.ID(),
			ScriptKey: inAsset.ScriptPubkey,
		}
		inputSet[*prevID] = inAsset

		rootAsset.PrevWitnesses[idx].PrevID = prevID
	}

	splitRoot, err := splitTree.Root(context.TODO())
	if err != nil {
		log.Println("splitRoot, err := splitTree.Root(context.TODO())", err)

		return nil, err
	}

	rootAsset.SplitCommitmentRoot = mssmt.NewComputedNode(splitRoot.NodeHash(), splitRoot.NodeSum())
	if err != nil {
		return nil, err
	}

	for idx := range locators {
		locator := locators[idx]

		proof, err := splitTree.MerkleProof(ctx, locator.Hash())
		if err != nil {
			return nil, err
		}

		prevWitnesses := splitAssets[*locator].PrevWitnesses
		prevWitnesses[0].SplitCommitment = &asset.SplitCommitment{
			Proof:     *proof,
			RootAsset: *rootAsset,
		}
	}

	return &SplitCommitment{
		PrevAssets:  inputSet,
		RootAsset:   rootAsset,
		SplitAssets: splitAssets,
		Tree:        splitTree,
	}, nil
}
