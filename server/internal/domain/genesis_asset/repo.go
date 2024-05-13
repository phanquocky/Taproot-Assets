package genesis_asset

import (
	"github.com/quocky/taproot-asset/server/internal/domain/common"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"golang.org/x/net/context"
)

type RepoInterface interface {
	common.RepoInterface
	FindAvailableAssetsWithAmount(ctx context.Context, pubkey []byte) (utxoasset.ListAssetsResp, error)
}
