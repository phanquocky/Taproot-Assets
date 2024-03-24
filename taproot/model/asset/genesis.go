package asset

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/btcsuite/btcd/wire"
)

type Genesis struct {
	FirstPrevOut wire.OutPoint
	Name         string
	OutputIndex  uint32
}

func NewGenesis(firstPrevOut wire.OutPoint, name string, outputIndex uint32) Genesis {
	return Genesis{
		FirstPrevOut: firstPrevOut,
		Name:         name,
		OutputIndex:  outputIndex,
	}
}

func (g Genesis) ID() ID {
	tagHash := sha256.Sum256([]byte(g.Name))

	h := sha256.New()
	_ = wire.WriteOutPoint(h, 0, 0, &g.FirstPrevOut)
	_, _ = h.Write(tagHash[:])
	_ = binary.Write(h, binary.BigEndian, g.OutputIndex)
	return *(*ID)(h.Sum(nil))
}
