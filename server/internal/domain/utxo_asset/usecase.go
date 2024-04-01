package utxoasset

import (
	"golang.org/x/net/context"
)

type UseCaseInterface interface {
	GetUnspentAssetsById(
		ctx context.Context,
		assetID string,
		amount int32,
		pubKey []byte,
	) (interface{}, error)
}
