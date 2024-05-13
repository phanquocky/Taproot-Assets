package utxoasset

import (
	utxoassetsdk "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"golang.org/x/net/context"
)

type UseCaseInterface interface {
	GetUnspentAssetsById(
		ctx context.Context,
		assetID string,
		amount int32,
		pubKey []byte,
	) (*utxoassetsdk.UnspentAssetResp, error)
	ListAllAssetsWithAmount(ctx context.Context, pubkey []byte) (utxoassetsdk.ListAssetsResp, error)
}
