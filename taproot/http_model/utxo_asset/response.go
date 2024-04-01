package utxoasset

import (
	"github.com/quocky/taproot-asset/taproot/model/asset"
	manageutxo "github.com/quocky/taproot-asset/taproot/model/manage_utxo"
)

type UnspentAssetResp struct {
	GenesisAsset     asset.GenesisAsset
	UnspentOutpoints []*manageutxo.UnspentOutpoint
	GenesisPoint     asset.GenesisPoint
	InputFilesBytes  [][]byte
}
