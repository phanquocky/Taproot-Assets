package utxoasset

type UnspentAssetReq struct {
	AssetID string `json:"asset_id"`
	Amount  int32  `json:"amount"`
	PubKey  []byte `json:"pub_key"`
}
