package genesispoint

import (
	"github.com/quocky/taproot-asset/server/internal/domain/genesis"
	cmrepo "github.com/quocky/taproot-asset/server/internal/repo/common"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoMongo struct {
	*cmrepo.RepoMongo
}

func NewRepoMongo(
	db *mongo.Database,
) genesis.RepoInterface {
	return &RepoMongo{
		cmrepo.NewRepoMongo(db, "genesis_points"),
	}
}
