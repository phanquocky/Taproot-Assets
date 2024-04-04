package utxoasset

import (
	"github.com/quocky/taproot-asset/taproot/model/asset"
	assetoutpointmodel "github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
)

type UnspentAssetResp struct {
	GenesisAsset     asset.GenesisAsset
	UnspentOutpoints []*assetoutpointmodel.UnspentOutpoint
	GenesisPoint     asset.GenesisPoint
	InputFilesBytes  [][]byte
}
