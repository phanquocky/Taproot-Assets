package asset

import "github.com/btcsuite/btcd/wire"

type ID [32]byte

type PrevID struct {
	OutPoint wire.OutPoint

	// ID is the asset ID of the previous asset tree.
	ID        ID
	ScriptKey SerializedKey
}

type Witness struct {
	PrevID *PrevID

	SplitCommitment *SplitCommitment
}

// IsSplitCommitWitness returns true if the witness is a split-commitment
// witness.
func IsSplitCommitWitness(witness Witness) bool {
	return witness.PrevID != nil &&
		witness.SplitCommitment != nil
}
