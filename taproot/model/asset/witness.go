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
