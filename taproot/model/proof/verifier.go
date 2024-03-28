package proof

import (
	"context"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

// Verify verifies the proof by ensuring that:
//
// 1. A valid inclusion proof for the resulting asset is included.
// 2. A valid inclusion proof for the split root, if the resulting asset
// is a split asset.
// 3. A set of valid exclusion Proofs for the resulting asset are
// included.
// 4. If this is a genesis asset, start by verifying the
// genesis reveal, which should be present for genesis assets.
// Non-genesis assets must not have a genesis or meta reveal.
// 5. Either a set of asset inputs with valid witnesses is included that
// satisfy the resulting state transition or a challenge witness is
// provided as part of an ownership proof.
func (p *Proof) Verify(
	ctx context.Context,
	prev *AssetSnapshot,
) (*AssetSnapshot, error) {

	// TODO: validate p.asset (check asset name)

	// 1. A valid inclusion proof for the resulting asset is included.
	assetCommitment, err := p.verifyInclusionProof()
	if err != nil {
		return nil, err
	}

	fmt.Println("Asset Commitment: ", assetCommitment)

	// 2. A valid inclusion proof for the split root, if the resulting asset
	// is a split asset.
	if p.Asset.HasSplitCommitmentWitness() {
		if p.SplitRootProof == nil {
			return nil, ErrMissingSplitRootProof
		}

		// TODO: missing split proof.
		if err := p.verifySplitRootProof(); err != nil {
			return nil, err
		}
	}

	// 3. A set of valid exclusion Proofs for the resulting asset are
	// included.
	if err := p.verifyExclusionProofs(); err != nil {
		return nil, err
	}

	// 4. If this is a genesis asset, start by verifying the
	// genesis reveal, which should be present for genesis assets.
	// Non-genesis assets must not have a genesis or meta reveal.
	isGenesisAsset := p.Asset.IsGenesisAsset()
	hasGenesisReveal := p.GenesisReveal != nil

	switch {
	case !isGenesisAsset && hasGenesisReveal:
		return nil, ErrNonGenesisAssetWithGenesisReveal
	case isGenesisAsset && !hasGenesisReveal:
		return nil, ErrGenesisRevealRequired
	case isGenesisAsset && hasGenesisReveal:
		if err := p.verifyGenesisReveal(); err != nil {
			return nil, err
		}
	}

	// 5. Either a set of asset inputs with valid witnesses is included that
	// satisfy the resulting state transition or a challenge witness is
	// provided as part of an ownership proof.
	var splitAsset bool = false

	return &AssetSnapshot{
		Asset: &p.Asset,
		OutPoint: wire.OutPoint{
			Hash:  p.AnchorTx.TxHash(),
			Index: p.InclusionProof.OutputIndex,
		},
		// AnchorBlockHash:   p.BlockHeader.BlockHash(),
		// AnchorBlockHeight: p.BlockHeight,
		AnchorTx:    &p.AnchorTx,
		OutputIndex: p.InclusionProof.OutputIndex,
		InternalKey: p.InclusionProof.InternalKey,
		ScriptRoot:  assetCommitment,
		SplitAsset:  splitAsset, // TODO: genesis process -> no-existed Split Asset
		// TapscriptSibling: tapscriptPreimage,
		// MetaReveal:       p.MetaReveal,
	}, nil
}

// verifyGenesisReveal checks that the genesis reveal present in the proof at
// minting validates against the asset ID and proof details.
func (p *Proof) verifyGenesisReveal() error {
	reveal := p.GenesisReveal

	if reveal == nil {
		return ErrGenesisRevealRequired
	}

	// Make sure the genesis reveal is consistent with the TLV fields in
	// the state transition proof.
	if reveal.FirstPrevOut != p.PrevOut {
		return ErrGenesisRevealPrevOutMismatch
	}

	if reveal.OutputIndex != p.InclusionProof.OutputIndex {
		return ErrGenesisRevealOutputIndexMismatch
	}

	// The genesis reveal determines the ID of an asset, so make sure it is
	// consistent. Since the asset ID commits to all fields of the genesis,
	// this is equivalent to checking equality for the genesis tag and type
	// fields that have not yet been verified.
	assetID := p.Asset.ID()
	if reveal.ID() != assetID {
		return ErrGenesisRevealAssetIDMismatch
	}

	return nil
}

// verifyExclusionProofs verifies all ExclusionProofs are valid.
func (p *Proof) verifyExclusionProofs() error {
	// Gather all P2TR outputs in the on-chain transaction.
	p2trOutputs := make(map[uint32]struct{})
	for i, txOut := range p.AnchorTx.TxOut {
		if uint32(i) == p.InclusionProof.OutputIndex {
			continue
		}

		if txscript.IsPayToTaproot(txOut.PkScript) {
			p2trOutputs[uint32(i)] = struct{}{}
		}
	}

	// Verify all the encoded exclusion Proofs.
	for i := range p.ExclusionProofs {
		exclusionProof := p.ExclusionProofs[i]

		_, err := verifyTaprootProof(
			&p.AnchorTx, exclusionProof, &p.Asset, false,
		)
		if err != nil {
			return err
		}

		delete(p2trOutputs, exclusionProof.OutputIndex)
	}

	// If any outputs are missing a proof, fail.
	if len(p2trOutputs) > 0 {
		return ErrMissingExclusionProofs
	}

	return nil
}

// verifyInclusionProof verifies the InclusionProof is valid.
func (p *Proof) verifyInclusionProof() (*commitment.TapCommitment, error) {
	return verifyTaprootProof(
		&p.AnchorTx, &p.InclusionProof, &p.Asset, true,
	)
}

// verifySplitRootProof verifies the SplitRootProof is valid.
func (p *Proof) verifySplitRootProof() error {
	rootAsset := &p.Asset.PrevWitnesses[0].SplitCommitment.RootAsset
	_, err := verifyTaprootProof(
		&p.AnchorTx, p.SplitRootProof, rootAsset, true,
	)

	return err
}

// verifyTaprootProof attempts to verify a TaprootProof for inclusion or
// exclusion of an asset. If the taproot proof was an inclusion proof, then the
// AssetCommitment is returned as well.
func verifyTaprootProof(
	anchor *wire.MsgTx,
	proof *TaprootProof,
	asset *asset.Asset,
	inclusion bool,
) (*commitment.TapCommitment, error) {
	// Extract the final taproot key from the output including/excluding the
	// asset, which we'll use to compare our derived key against.
	expectedTaprootKey, err := ExtractTaprootKey(
		anchor, proof.OutputIndex,
	)
	if err != nil {
		return nil, err
	}

	// For each proof type, we'll map this to a single key based on the
	// self-identified pre-image type in the specified proof.
	var (
		derivedKey    *btcec.PublicKey
		tapCommitment *commitment.TapCommitment
	)
	switch {
	// If this is an inclusion proof, then we'll derive the expected
	// taproot output key based on the revealed asset MS-SMT proof. The
	// root of this tree will then be used to assemble the top of the
	// tapscript tree, which will then be tweaked as normal with the
	// internal key to derive the expected output key.
	case inclusion:
		log.Printf("Verifying inclusion proof for asset with id %x \n", asset.ID())

		derivedKey, tapCommitment, err = proof.DeriveByAssetInclusion(
			asset,
		)

	// If the commitment proof is present, then this is actually a
	// non-inclusion proof: we want to verify that either no root
	// commitment exists, or one does, but the asset in question isn't
	// present.
	case proof.CommitmentProof != nil:
		log.Printf("Verifying exclusion proof for asset with id %x \n", asset.ID())
		derivedKey, err = proof.DeriveByAssetExclusion(
			asset.AssetCommitmentKey(),
			asset.TapCommitmentKey(),
		)

	// If this is a tapscript proof, then we want to verify that the target
	// output DOES NOT contain any sort of Taproot Asset commitment.
	case proof.TapscriptProof != nil:
		log.Println("Verifying tapscript proof")
		derivedKey, err = proof.DeriveByTapscriptProof()
	}
	if err != nil {
		return nil, err
	}

	log.Println("Derived key: ", derivedKey)
	log.Println("Expected key: ", expectedTaprootKey)
	// The derived key should match the extracted key.
	if derivedKey.IsEqual(expectedTaprootKey) {
		return tapCommitment, nil
	}

	return nil, commitment.ErrInvalidTaprootProof
}

// Verify attempts to verify a full proof file starting from the asset's
// genesis.
//
// The passed context can be used to exit early from the inner proof
// verification loop.
//
// TODO(roasbeef): pass in the expected genesis point here?
func (f *File) Verify(ctx context.Context) (*AssetSnapshot, error) {
	var prev *AssetSnapshot
	for idx := range f.Proofs {
		decodedProof, err := f.ProofAt(uint32(idx))
		if err != nil {
			return nil, err
		}

		result, err := decodedProof.Verify(ctx, prev)
		if err != nil {
			return nil, err
		}
		prev = result
	}

	return prev, nil
}
