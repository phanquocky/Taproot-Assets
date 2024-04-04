package assetoutpoint

import "github.com/quocky/taproot-asset/server/internal/domain/common"

type UnspentOutpointFilter struct {
	IDs       *common.InOperator `json:"_id,omitempty"`
	GenesisID *common.ID         `json:"genesis_id,omitempty"`
	Spent     *bool              `json:"spent,omitempty"`
	ScriptKey []byte             `json:"script_key,omitempty"`
}

type UnspentOutpointUpdate struct {
	Set *UnspentOutpointSetUpdate `json:"$set,omitempty"`
}

type UnspentOutpointSetUpdate struct {
	Spent *bool `json:"spent,omitempty"`
}
