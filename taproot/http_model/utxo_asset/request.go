package utxoasset

type UnspentAssetReq struct {
	AssetID string `json:"asset_id"`
	Amount  int32  `json:"amount"`
	PubKey  []byte `json:"pub_key"`
}

type ListAssetReq struct {
	Pubkey []byte `json:"pub_key"`
}

type ListAssetResp struct {
	Amount  int32  `json:"amount"`
	Name    string `json:"asset_name"`
	AssetID []byte `json:"asset_id"`
}

type ListAssetsResp []*ListAssetResp
