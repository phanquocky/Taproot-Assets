package proof

import (
	"context"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/utils"
	"log"
)

type AssetProofs map[asset.SerializedKey]*Proof

type MintParams struct {
	BaseProofParams
	GenesisPoint wire.OutPoint
}

// NewMintingBlobs takes a set of minting parameters, and produces a series of
// serialized proof files, which proves the creation/existence of each of the
// assets within transaction onchain.
func NewMintingBlobs(params *MintParams) (AssetProofs, error) {
	base, err := baseProof(&params.BaseProofParams, params.GenesisPoint)
	if err != nil {
		return nil, err
	}

	proofs, err := committedProofs(
		base, params.TaprootAssetRoot,
	)
	if err != nil {
		log.Println("proof, err := committedProofs(", err)

		return nil, err
	}

	fmt.Println("proofs: ", proofs)

	ctx := context.Background()

	for key := range proofs {
		fmt.Println("key: ", key)
		proof := proofs[key]

		// Verify the generated Proofs.
		_, err = proof.Verify(ctx, nil)

		log.Println("send proof to BADUY")
		utils.PrintStruct(proof)

		if err != nil {
			return nil, fmt.Errorf("invalid proof file generated: %w", err)
		}
		if err == nil {
			fmt.Println("Verify done!!!")
		}
	}

	return proofs, nil
}

// baseProof creates the basic proof template that contains all anchor
// transaction related fields.
func baseProof(params *BaseProofParams, prevOut wire.OutPoint) (*Proof, error) {
	// First, we'll create the merkle proof for the anchor transaction. In
	// this case, since all the assets were created in the same block, we
	// only need a single merkle proof.
	proof, err := coreProof(params)
	if err != nil {
		return nil, err
	}

	// Now, we'll construct the base proof that all the assets created in
	// this batch or spent in this transaction will share.
	proof.PrevOut = prevOut
	proof.InclusionProof = TaprootProof{
		OutputIndex: uint32(params.OutputIndex),
		InternalKey: params.InternalKey,
	}
	proof.ExclusionProofs = params.ExclusionProofs
	return proof, nil
}

// coreProof creates the basic proof template that contains only fields
// dependent on anchor transaction confirmation.
func coreProof(params *BaseProofParams) (*Proof, error) {
	return &Proof{
		AnchorTx: *params.Tx,
	}, nil
}

// committedProofs creates a proof of the given params.
func committedProofs(baseProof *Proof, tapCommitment *commitment.TapCommitment,
) (AssetProofs, error) {

	commitAssets := tapCommitment.CommittedAssets()
	proofs := make(AssetProofs, len(commitAssets))
	fmt.Println("commitAssets: ", commitAssets)
	for i, _ := range commitAssets {
		newAsset := commitAssets[i]

		assetProof := *baseProof
		assetProof.Asset = *newAsset.Copy()

		// With the base information contained, we'll now need to
		// generate our series of MS-SMT inclusion Proofs that prove
		// the existence of the asset.
		_, assetMerkleProof, err := tapCommitment.Proof(
			newAsset.TapCommitmentKey(),
			newAsset.AssetCommitmentKey(),
		) // TODO????
		if err != nil {
			return nil, err
		}

		fmt.Println("assetMerkleProof123345: ", assetMerkleProof)

		assetProof.InclusionProof.CommitmentProof = &commitment.CommitmentProof{
			AssetProof:        assetMerkleProof.AssetProof,
			TaprootAssetProof: assetMerkleProof.TaprootAssetProof,
		}

		// Set the genesis reveal info on the minting proof. To save on
		// some space, the genesis info is no longer included in
		// transition Proofs.
		assetProof.GenesisReveal = &newAsset.Genesis

		pubkey, _ := newAsset.ScriptPubkey.ToPubKey()
		scriptKey := asset.ToSerialized(pubkey)

		fmt.Println("scriptKey123: ", i, scriptKey)
		proofs[scriptKey] = &assetProof

		fmt.Println("newAsset: ", i, newAsset)
		fmt.Println("scriptKey: ", i, scriptKey)
	}
	fmt.Println("proofs: ", proofs)
	return proofs, nil
}
