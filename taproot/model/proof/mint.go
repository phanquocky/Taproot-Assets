package proof

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"

	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

type AssetProofs []*Proof

type MintParams struct {
	BaseProofParams
	GenesisPoint wire.OutPoint
}

func NewMintingBlobs(logger *zap.Logger, params *MintParams) (AssetProofs, error) {
	template := createTemplateProof(&params.BaseProofParams, params.GenesisPoint)

	logger.Debug("template proof: ", zap.Reflect("template", template))
	proofs, err := committedProofs(
		template, params.TapCommitment,
	)
	if err != nil {
		return nil, err
	}

	logger.Debug("committed proofs: ", zap.Reflect("proofs", proofs))

	ctx := context.Background()

	for key := range proofs {
		proof := proofs[key]

		_, err = proof.Verify(ctx, nil)

		if err != nil {
			return nil, fmt.Errorf("generate invalid proof: %w", err)
		}
	}

	log.Println("Verify proof success!")

	return proofs, nil
}

func createTemplateProof(params *BaseProofParams, genesisPoint wire.OutPoint) *Proof {
	proof := &Proof{
		AnchorTx: *params.Tx,
	}

	proof.PrevOut = genesisPoint
	proof.InclusionProof = TaprootProof{
		OutputIndex: uint32(params.OutputIndex),
		InternalKey: params.InternalKey,
	}
	proof.ExclusionProofs = params.ExclusionProofs
	return proof
}

//func coreProof(params *BaseProofParams) (*CreateProof, error) {
//	return &CreateProof{
//		AnchorTx: *params.Tx,
//	}, nil
//}

func committedProofs(template *Proof, tapCommitment *commitment.TapCommitment) (AssetProofs, error) {

	assets := tapCommitment.Assets()
	proofs := make(AssetProofs, len(assets))

	for i, _ := range assets {
		newAsset := assets[i]

		assetProof := *template
		assetProof.Asset = *newAsset.Copy()

		_, commitmentProof, err := tapCommitment.CreateProof(
			newAsset.TapCommitmentKey(),
			newAsset.AssetCommitmentKey(),
		)
		if err != nil {
			return nil, err
		}

		assetProof.InclusionProof.CommitmentProof = commitmentProof
		assetProof.GenesisReveal = &newAsset.Genesis

		proofs[i] = &assetProof
	}

	return proofs, nil
}
