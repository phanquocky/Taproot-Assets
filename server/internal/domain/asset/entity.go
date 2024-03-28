package asset

import "github.com/quocky/taproot-asset/server/internal/domain/common"

type GenesisAsset struct {
	common.Entity  `json:",inline"`
	AssetID        []byte `json:"asset_id"`
	AssetName      string `json:"asset_name"`
	Supply         int32  `json:"supply"`
	OutputIndex    int32  `json:"output_index"`
	GenesisPointID int32  `json:"genesis_point_id"`
}
