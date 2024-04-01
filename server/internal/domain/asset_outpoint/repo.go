package assetoutpoint

import (
	"github.com/quocky/taproot-asset/server/internal/domain/common"
	"golang.org/x/net/context"
)

type RepoInterface interface {
	common.RepoInterface
	FindManyWithManagedUTXO(ctx context.Context, filter any) ([]*UnspentOutpoint, error)
}
