package manageutxo

import "github.com/quocky/taproot-asset/server/internal/domain/common"

type ManagedUtxo struct {
	common.Entity    `json:",inline"`
	Outpoint         string    `json:"outpoint"`
	AmtSats          int32     `json:"amt_sats"`
	InternalKey      []byte    `json:"internal_key"`
	TaprootAssetRoot []byte    `json:"taproot_asset_root"`
	ScriptOutput     []byte    `json:"script_output"`
	TxID             common.ID `json:"tx_id"`
}
