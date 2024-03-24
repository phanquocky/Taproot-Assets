package proof

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

type BaseProofParams struct {
	Tx               *wire.MsgTx
	OutputIndex      int32
	InternalKey      asset.SerializedKey
	TaprootAssetRoot *commitment.TapCommitment
	ExclusionProofs  []*TaprootProof
}

func (b *BaseProofParams) AddExclusionProofs(
	txIncludeOutPubKey *onchain.TxIncludeOutInternalKey,
	isAnchor func(uint32) bool,
) error {
	tx := txIncludeOutPubKey.Tx

	for outIdx := range tx.TxOut {
		txOut := tx.TxOut[outIdx]

		// Skip any anchor output since that will get an inclusion proof
		// instead.
		if isAnchor(uint32(outIdx)) {
			continue
		}

		// We only need to add exclusion Proofs for P2TR outputs as only
		// those could commit to a Taproot Asset tree.
		if !txscript.IsPayToTaproot(txOut.PkScript) {
			continue
		}

		// For a P2TR output the internal key must be declared and must
		// be a valid 32-byte x-only public key.
		internalKey := txIncludeOutPubKey.OutInternalKeys[outIdx]

		// Okay, we now know this is a normal BIP-0086 key spend and can
		// add the exclusion proof accordingly.
		b.ExclusionProofs = append(
			b.ExclusionProofs, &TaprootProof{
				OutputIndex: uint32(outIdx),
				InternalKey: internalKey,
				TapscriptProof: &TapscriptProof{
					Bip86: true,
				},
			},
		)
	}

	return nil
}
