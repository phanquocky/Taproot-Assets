package proof

import (
	"errors"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
)

var (
	// ErrMissingSplitRootProof is an error returned upon noticing an
	// inclusion proof for a split root asset is missing.
	ErrMissingSplitRootProof = errors.New("missing split root proof")

	// ErrMissingExclusionProofs is an error returned upon noticing an
	// exclusion proof for a P2TR output is missing.
	ErrMissingExclusionProofs = errors.New("missing exclusion proof(s)")

	// ErrNonGenesisAssetWithGenesisReveal is an error returned if an asset
	// proof for a non-genesis asset contains a genesis reveal.
	ErrNonGenesisAssetWithGenesisReveal = errors.New("non genesis asset " +
		"has genesis reveal")

	// ErrGenesisRevealRequired is an error returned if an asset proof for a
	// genesis asset is missing a genesis reveal.
	ErrGenesisRevealRequired = errors.New("genesis reveal required")

	// ErrGenesisRevealPrevOutMismatch is an error returned if an asset
	// proof for a genesis asset has a genesis reveal where the prev out
	// doesn't match the proof TLV field.
	ErrGenesisRevealPrevOutMismatch = errors.New("genesis reveal prev " +
		"out mismatch")

	// ErrGenesisRevealOutputIndexMismatch is an error returned if an asset
	// proof for a genesis asset has a genesis reveal where the output index
	// doesn't match the proof TLV field.
	ErrGenesisRevealOutputIndexMismatch = errors.New("genesis reveal " +
		"output index mismatch")

	// ErrGenesisRevealAssetIDMismatch is an error returned if an asset
	// proof for a genesis asset has a genesis reveal that is inconsistent
	// with the asset ID.
	ErrGenesisRevealAssetIDMismatch = errors.New("genesis reveal asset " +
		"ID mismatch")
)

type Interface interface {
	Verify() (bool, error)
	Store() error
}

type Proof struct {
	PrevOut          wire.OutPoint // TODO: genesisPoint
	AnchorTx         wire.MsgTx
	Asset            asset.Asset
	InclusionProof   TaprootProof
	ExclusionProofs  []*TaprootProof
	SplitRootProof   *TaprootProof
	AdditionalInputs []File
	GenesisReveal    *asset.Genesis
}
