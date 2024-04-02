package proof

import (
	"context"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

func (p *Proof) Verify(
	ctx context.Context,
	prev *AssetSnapshot,
) (*AssetSnapshot, error) {

	// TODO: validate p.asset (check asset name)

	assetCommitment, err := p.verifyInclusionProof()
	if err != nil {
		return nil, err
	}

	if p.Asset.HasSplitCommitmentWitness() {
		if p.SplitRootProof == nil {
			return nil, ErrMissingSplitRootProof
		}

		if err := p.verifySplitRootProof(); err != nil {
			return nil, err
		}
	}

	if err := p.verifyExclusionProofs(); err != nil {
		return nil, err
	}

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

	var splitAsset bool = false

	return &AssetSnapshot{
		Asset: &p.Asset,
		OutPoint: wire.OutPoint{
			Hash:  p.AnchorTx.TxHash(),
			Index: p.InclusionProof.OutputIndex,
		},
		AnchorTx:    &p.AnchorTx,
		OutputIndex: p.InclusionProof.OutputIndex,
		InternalKey: p.InclusionProof.InternalKey,
		ScriptRoot:  assetCommitment,
		SplitAsset:  splitAsset, // TODO: genesis process -> no-existed Split Asset
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

func (p *Proof) verifyInclusionProof() (*commitment.TapCommitment, error) {
	return verifyTaprootProof(
		&p.AnchorTx, &p.InclusionProof, &p.Asset, true,
	)
}

func (p *Proof) verifySplitRootProof() error {
	rootAsset := &p.Asset.PrevWitnesses[0].SplitCommitment.RootAsset
	_, err := verifyTaprootProof(
		&p.AnchorTx, p.SplitRootProof, rootAsset, true,
	)

	return err
}

func verifyTaprootProof(
	anchor *wire.MsgTx,
	proof *TaprootProof,
	asset *asset.Asset,
	inclusion bool,
) (*commitment.TapCommitment, error) {

	expectedTaprootKey, err := ExtractTaprootKey(
		anchor, proof.OutputIndex,
	)
	if err != nil {
		return nil, err
	}

	var (
		derivedKey    *btcec.PublicKey
		tapCommitment *commitment.TapCommitment
	)
	switch {
	case inclusion:
		log.Printf("Verifying inclusion proof for asset with id %x \n", asset.ID())

		derivedKey, tapCommitment, err = proof.DeriveByAssetInclusion(
			asset,
		)

	case proof.CommitmentProof != nil:
		log.Printf("Verifying exclusion proof for asset with id %x \n", asset.ID())
		derivedKey, err = proof.DeriveByAssetExclusion(
			asset.AssetCommitmentKey(),
			asset.TapCommitmentKey(),
		)

	case proof.TapscriptProof != nil:
		log.Println("Verifying tapscript proof")
		derivedKey, err = proof.DeriveByTapscriptProof()
	}
	if err != nil {
		return nil, err
	}

	if derivedKey.IsEqual(expectedTaprootKey) {
		return tapCommitment, nil
	}

	return nil, commitment.ErrInvalidTaprootProof
}

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
