package chaintx

import "github.com/quocky/taproot-asset/server/internal/domain/common"

type ChainTx struct {
	common.Entity `json:",inline"`
	TxID          []byte `json:"tx_id"`
	AnchorTx      []byte `json:"anchor_tx"`
	ChainFees     int64  `json:"chain_fees,omitempty"`
	BlockHeight   int32  `json:"block_height,omitempty"`
	BlockHash     []byte `json:"block_hash,omitempty"`
}
