package assetoutpoint

import (
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	cmrepo "github.com/quocky/taproot-asset/server/internal/repo/common"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoMongo struct {
	*cmrepo.RepoMongo
}

func NewRepoMongo(
	db *mongo.Database,
) assetoutpoint.RepoInterface {
	return &RepoMongo{
		cmrepo.NewRepoMongo(db, "asset_outpoints"),
	}
}
