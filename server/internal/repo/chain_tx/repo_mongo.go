package chaintx

import (
	"github.com/quocky/taproot-asset/server/internal/domain/chain_tx"
	cmrepo "github.com/quocky/taproot-asset/server/internal/repo/common"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoMongo struct {
	*cmrepo.RepoMongo
}

func NewRepoMongo(
	db *mongo.Database,
) chaintx.RepoInterface {
	return &RepoMongo{
		cmrepo.NewRepoMongo(db, "chain_txs"),
	}
}
