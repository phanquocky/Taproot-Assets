package proof

import (
	"context"
	"fmt"

	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/utils"

	"github.com/btcsuite/btcd/wire"
)

// TransitionParams holds the set of chain level information needed to append a
// proof to an existing file for the given asset state transition.
type TransitionParams struct {
	// BaseProofParams houses the basic chain level parameters needed to
	// construct a proof.
	BaseProofParams

	// NewAsset is the new asset created by the asset transition.
	NewAsset *asset.Asset

	// RootOutputIndex is the index of the output that commits to the split
	// root asset, if present.
	RootOutputIndex uint32

	// RootInternalKey is the internal key of the output at RootOutputIndex.
	RootInternalKey asset.SerializedKey

	// RootTaprootAssetTree is the commitment root that commitments to the
	// inclusion of the root split asset at the RootOutputIndex.
	RootTaprootAssetTree *commitment.TapCommitment
}

// AppendTransition appends a new proof for a state transition to the given
// encoded proof file. Because multiple assets can be committed to in the same
// on-chain output, this function takes the script key of the asset to return
// the proof for. This method returns both the encoded full provenance (proof
// chain) and the added latest proof.
func AppendTransition(
	inputFilesBytes [][]byte,
	params *TransitionParams,
) (*File, *Proof, error) {
	ctx := context.Background()

	// Decode the proof blob into a proper file structure first.
	f := File{}
	if err := f.Decode(inputFilesBytes[0]); err != nil {
		return nil, nil, fmt.Errorf("error decoding proof file: %w",
			err)
	}

	// Cannot add a transition to an empty proof file.
	if f.IsEmpty() {
		return nil, nil, fmt.Errorf("invalid empty proof file")
	}

	lastProof, err := f.LastProof()
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching last proof: %w",
			err)
	}

	_, err = lastProof.Verify(context.TODO(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error verifying last proof: %w", err)
	}

	lastPrevOut := wire.OutPoint{
		Hash:  lastProof.AnchorTx.TxHash(),
		Index: lastProof.InclusionProof.OutputIndex,
	}

	// We can now create the new proof entry for the asset in the params.
	newProof, err := CreateTransitionProof(lastPrevOut, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating transition "+
			"proof: %w", err)
	}

	if len(inputFilesBytes) > 1 {
		additionalFiles := make([]File, len(inputFilesBytes)-1)

		for i, inputFilesByte := range inputFilesBytes[1:] {
			fInput := File{}
			if err := fInput.Decode(inputFilesByte); err != nil {
				return nil, nil, fmt.Errorf("error decoding proof file: %w",
					err)
			}

			if _, err := fInput.Verify(ctx); err != nil {
				return nil, nil, fmt.Errorf("error verify file: %w", err)
			}

			additionalFiles[i] = fInput
		}

		newProof.AdditionalInputs = additionalFiles
	}

	if _, err := newProof.Verify(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("error verifying proof new proof: %w", err)
	}

	// Before we encode and return the proof, we want to validate it. For
	// that we need to start at the beginning.
	if err := f.AppendProof(*newProof); err != nil {
		return nil, nil, fmt.Errorf("error appending proof: %w", err)
	}

	if _, err := f.Verify(ctx); err != nil {
		return nil, nil, fmt.Errorf("error verifying proof: %w", err)
	}

	return &f, newProof, nil
}

// CreateTransitionProof creates a proof for an asset transition, based on the
// last proof of the last asset state and the new asset in the params.
func CreateTransitionProof(prevOut wire.OutPoint,
	params *TransitionParams) (*Proof, error) {

	proof := createTemplateProof(&params.BaseProofParams, prevOut)

	proof.Asset = *params.NewAsset.Copy()
	proof.Asset.PrevWitnesses[0].SplitCommitment = nil

	fmt.Println("proof.Assetproof.Assetproof.Assetproof.Assetproof.Asset")
	utils.PrintStruct(proof.Asset.PrevWitnesses[0].PrevID)

	_, assetMerkleProof, err := params.TapCommitment.CreateProof(
		proof.Asset.TapCommitmentKey(),
		proof.Asset.AssetCommitmentKey(),
	)
	if err != nil {
		return nil, err
	}

	proof.InclusionProof.CommitmentProof = assetMerkleProof
	//
	//if proof.Asset.HasSplitCommitmentWitness() {
	//	splitAsset := proof.Asset
	//	rootAsset := &splitAsset.PrevWitnesses[0].SplitCommitment.RootAsset
	//
	//	rootTree := params.RootTaprootAssetTree
	//	committedRoot, rootMerkleProof, err := rootTree.CreateProof(
	//		rootAsset.TapCommitmentKey(),
	//		rootAsset.AssetCommitmentKey(),
	//	)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	// If the asset wasn't committed to, the proof is invalid.
	//	if committedRoot == nil {
	//		return nil, fmt.Errorf("no asset commitment found")
	//	}
	//
	//	// Make sure the committed asset matches the root asset exactly.
	//	if !committedRoot.DeepEqual(rootAsset) {
	//		return nil, fmt.Errorf("root asset mismatch")
	//	}
	//
	//	proof.SplitRootProof = &TaprootProof{
	//		OutputIndex:     params.RootOutputIndex,
	//		InternalKey:     params.RootInternalKey,
	//		CommitmentProof: rootMerkleProof,
	//	}
	//}

	return proof, nil
}
