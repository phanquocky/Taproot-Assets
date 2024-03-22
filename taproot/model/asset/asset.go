package asset

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
)

// SerializedKey compress public key to 33 bytes
type SerializedKey [33]byte

type Genesis struct {
	FirstPrevOut wire.OutPoint
	Name         string
	OutputIndex  uint32
}

type Asset struct {
	Genesis
	Amount              uint64
	ScriptPubkey        SerializedKey
	SplitCommitmentRoot *mssmt.ComputedNode
	PrevWitnesses       []Witness
}
