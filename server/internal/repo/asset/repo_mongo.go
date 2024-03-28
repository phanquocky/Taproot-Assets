package asset

import (
	"github.com/quocky/taproot-asset/server/internal/domain/asset"
	cmrepo "github.com/quocky/taproot-asset/server/internal/repo/common"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoMongo struct {
	*cmrepo.RepoMongo
}

func NewRepoMongo(
	db *mongo.Database,
) asset.RepoInterface {
	return &RepoMongo{
		cmrepo.NewRepoMongo(db, "assets"),
	}
}
