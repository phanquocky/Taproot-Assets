package mint

import "github.com/quocky/taproot-asset/taproot/model/asset"

type MintAssetsResp []MintAssetResp

type MintAssetResp struct {
	ID     asset.ID `json:"asset_id"`
	Name   string   `json:"asset_name"`
	Amount int32    `json:"amount"`
}
