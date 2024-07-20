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

// func (g Genesis) String() string {
// 	return fmt.Sprintf("Genesis{FirstPrevOut: %s, Name: %s, OutputIndex: %d}", g.FirstPrevOut, g.Name, g.OutputIndex)
// }

type GenesisAsset struct {
	// AssetID have to use hex.Encoder to convert to []byte
	AssetID        []byte `json:"asset_id"`
	AssetName      string `json:"asset_name"`
	Supply         int32  `json:"supply"`
	OutputIndex    int32  `json:"output_index"`
	GenesisPointID string `json:"genesis_point_id"`
}

type GenesisPoint struct {
	PrevOut    string `json:"prev_out"`
	AnchorTxID string `json:"anchor_tx_id"`
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
