package assetoutpoint

import "github.com/quocky/taproot-asset/server/internal/domain/common"

type UnspentOutpointFilter struct {
	GenesisID *common.ID `json:"genesis_id,omitempty"`
	Spent     *bool      `json:"spent,omitempty"`
	ScriptKey []byte     `json:"script_key,omitempty"`
}
