package genesis

import (
	"github.com/quocky/taproot-asset/server/internal/domain/common"
)

type GenesisPoint struct {
	common.Entity `json:",inline"`
	PrevOut       string `json:"prev_out"`
	AnchorTxID    int32  `json:"anchor_tx_id"`
}
