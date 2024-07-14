package proof

import (
	"errors"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

var (
	// ErrInvalidCommitmentProof is an error returned upon attempting to
	// prove a malformed CommitmentProof.
	ErrInvalidCommitmentProof = errors.New(
		"invalid Taproot Asset commitment proof",
	)
)

type TapscriptProof struct {
	Bip86 bool
}

type TaprootProof struct {
	OutputIndex     uint32
	InternalKey     asset.SerializedKey
	CommitmentProof *commitment.CommitmentProof
	TapscriptProof  *TapscriptProof
}

// DeriveTaprootKeys derives the expected taproot key from a TapscriptProof
// backing a taproot output that does not include a Taproot Asset commitment.
//
// There are at most two possible keys to try if each leaf preimage matches the
// length of a branch preimage. However, based on the annotated type
// information, we only need to derive a single expected key.
func (p TapscriptProof) DeriveTaprootKeys(internalKey asset.SerializedKey) (
	*btcec.PublicKey,
	error,
) {
	tapscriptRoot := []byte{}

	pubkey, err := internalKey.ToPubKey()
	if err != nil {
		log.Println("[DeriveTaprootKeys] convert SerializedKey to PubKey fail", err)

		return nil, err
	}

	// Now that we have the expected tapscript root, we'll derive our
	// expected tapscript root.
	taprootKey := txscript.ComputeTaprootOutputKey(
		pubkey, tapscriptRoot,
	)

	// TODO(roasbeef): same here -- just need to verify as actual
	// control block proof?
	return schnorr.ParsePubKey(schnorr.SerializePubKey(taprootKey))
}

// DeriveByAssetExclusion derives the possible taproot keys backing a Taproot
// Asset commitment by interpreting the TaprootProof as an asset exclusion
// proof. Asset exclusion Proofs can take two forms: one where an asset proof
// proves that the asset no longer exists within its AssetCommitment, and
// another where the AssetCommitment corresponding to the excluded asset no
// longer exists within the TapCommitment.
//
// There are at most two possible keys to try if each leaf preimage matches the
// length of a branch preimage. However, based on the type of the sibling
// pre-image we'll derive just a single version of it.
func (p TaprootProof) DeriveByAssetExclusion(
	assetCommitmentKey,
	tapCommitmentKey [32]byte,
) (*btcec.PublicKey, error) {

	if p.CommitmentProof == nil || p.TapscriptProof != nil {
		return nil, ErrInvalidCommitmentProof
	}

	// Use the commitment proof to go from the empty asset leaf or empty
	// asset commitment leaf all the way up to the Taproot Asset commitment
	// root, which is then mapped to a TapLeaf and is hashed with a sibling
	// node, if any, to derive the tapscript root and taproot output key.
	// We'll do this twice, one for the possible branch sibling and another
	// for the possible leaf sibling.
	var (
		tapCommitment *commitment.TapCommitment
		err           error
	)

	switch {
	// In this case, there's no asset proof, so we want to verify that the
	// specified key maps to an empty leaf node (no asset ID sub-tree in
	// the root commitment).
	case p.CommitmentProof.AssetProof == nil:
		log.Printf("Deriving commitment by asset commitment exclusion")
		tapCommitment, err = p.CommitmentProof.
			DeriveByAssetCommitmentExclusion(tapCommitmentKey)

	// Otherwise, we have an asset proof, which means the tree contains the
	// asset ID, but we want to verify that the particular asset we care
	// about isn't included.
	default:
		log.Printf("Deriving commitment by asset exclusion")
		tapCommitment, err = p.CommitmentProof.
			DeriveByAssetExclusion(assetCommitmentKey)
	}
	if err != nil {
		return nil, err
	}

	pubkey, err := p.InternalKey.ToPubKey()
	if err != nil {
		log.Println("[DeriveByAssetInclusion] - convert SerializedKey to pubkey fail", err)

		return nil, err
	}

	return deriveTaprootKeysFromTapCommitment(
		tapCommitment, pubkey,
	)
}
func deriveTaprootKeysFromTapCommitment(commitment *commitment.TapCommitment,
	internalKey *btcec.PublicKey,
) (*btcec.PublicKey, error) {
	return deriveTaprootKeyFromTapCommitment(
		commitment, internalKey,
	)
}

// deriveTaprootKey derives the taproot key backing a Taproot Asset commitment.
func deriveTaprootKeyFromTapCommitment(
	tapCommitment *commitment.TapCommitment,
	internalKey *btcec.PublicKey,
) (*btcec.PublicKey, error) {

	commitmentLeaf := tapCommitment.TapLeaf()
	tapscriptRoot := txscript.AssembleTaprootScriptTree(commitmentLeaf).
		RootNode.TapHash()

	return schnorr.ParsePubKey(schnorr.SerializePubKey(
		txscript.ComputeTaprootOutputKey(internalKey, tapscriptRoot[:]),
	))
}

// DeriveByTapscriptProof derives the possible taproot keys from a
// TapscriptProof backing a taproot output that does not include a Taproot Asset
// commitment.
//
// NOTE: There are at most two possible keys to try if each leaf preimage
// matches the length of a branch preimage. However, we can derive only the one
// specified in the contained proof.
func (p TaprootProof) DeriveByTapscriptProof() (*btcec.PublicKey, error) {
	if p.CommitmentProof != nil || p.TapscriptProof == nil {

		return nil, commitment.ErrInvalidTapscriptProof
	}

	return p.TapscriptProof.DeriveTaprootKeys(p.InternalKey)
}

// DeriveByAssetInclusion derives the unique taproot output key backing a
// Taproot Asset commitment by interpreting the TaprootProof as an asset
// inclusion proof.
func (p TaprootProof) DeriveByAssetInclusion(
	asset *asset.Asset,
) (*btcec.PublicKey, *commitment.TapCommitment, error) {

	if p.CommitmentProof == nil || p.TapscriptProof != nil {
		return nil, nil, ErrInvalidCommitmentProof
	}

	// If this is an asset with a split commitment, then we need to verify
	// the inclusion proof without this information. As the output of the
	// receiver was created without this present.
	if asset.HasSplitCommitmentWitness() {
		asset = asset.Copy()
		asset.PrevWitnesses[0].SplitCommitment = nil
	}

	// Use the commitment proof to go from the asset leaf all the way up to
	// the Taproot Asset commitment root, which is then mapped to a TapLeaf
	// and is hashed with a sibling node, if any, to derive the tapscript
	// root and taproot output key.
	tapCommitment, err := p.CommitmentProof.DeriveByAssetInclusion(asset)
	if err != nil {
		return nil, nil, err
	}

	pubkey, err := p.InternalKey.ToPubKey()
	if err != nil {
		log.Println("[DeriveByAssetInclusion] - convert SerializedKey to pubkey fail", err)

		return nil, nil, err
	}

	pubKey, err := deriveTaprootKeyFromTapCommitment(
		tapCommitment, pubkey,
	)
	if err != nil {
		return nil, nil, err
	}

	return pubKey, tapCommitment, nil
}

// deriveTaprootKeyFromAssetCommitment derives the taproot key backing a Taproot Asset commitment.
func deriveTaprootKeyFromAssetCommitment(commitment *commitment.AssetCommitment,
	internalKey asset.SerializedKey) (
	*btcec.PublicKey, error) {

	// TODO(roasbeef): should just be control block proof verification?
	//  * should be getting the party bit from that itself
	commitmentLeaf := commitment.TapLeaf()

	tapscriptRoot := txscript.AssembleTaprootScriptTree(commitmentLeaf).RootNode.TapHash()

	pubkey, err := internalKey.ToPubKey()
	if err != nil {
		log.Println("convert SerializedKey to pubkey fail", err)

		return nil, err
	}

	return schnorr.ParsePubKey(schnorr.SerializePubKey(
		txscript.ComputeTaprootOutputKey(pubkey, tapscriptRoot[:]),
	))
}
