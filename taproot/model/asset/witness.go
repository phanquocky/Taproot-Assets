package asset

import (
	"github.com/btcsuite/btcd/wire"
	"reflect"
)

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

// DeepEqual returns true if this witness is equal with the given witness.
func (w *Witness) DeepEqual(o *Witness) bool {
	if w == nil || o == nil {
		return w == o
	}

	if !reflect.DeepEqual(w.PrevID, o.PrevID) {
		return false
	}

	return w.SplitCommitment.DeepEqual(o.SplitCommitment)
	//return w.SplitCommitment.DeepEqual(o.SplitCommitment)
}
