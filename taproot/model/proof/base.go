package proof

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

type BaseProofParams struct {
	Tx              *wire.MsgTx
	OutputIndex     int32
	InternalKey     asset.SerializedKey
	TapCommitment   *commitment.TapCommitment
	ExclusionProofs []*TaprootProof
}

func (b *BaseProofParams) AddExclusionProofs(
	txIncludeOutPubKey *onchain.TxIncludeOutPubKey,
	isAnchor func(int32) bool,
) error {
	tx := txIncludeOutPubKey.Tx

	for outIdx := range tx.TxOut {
		txOut := tx.TxOut[outIdx]

		if isAnchor(int32(outIdx)) {
			continue
		}

		if !txscript.IsPayToTaproot(txOut.PkScript) {
			continue
		}

		internalKey := txIncludeOutPubKey.OutPubKeys[int32(outIdx)]

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
