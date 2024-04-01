package assetoutpoint

import (
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	cmrepo "github.com/quocky/taproot-asset/server/internal/repo/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type RepoMongo struct {
	*cmrepo.RepoMongo
}

func (r *RepoMongo) FindManyWithManagedUTXO(
	ctx context.Context,
	filter any,
) ([]*assetoutpoint.UnspentOutpoint, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "manage_utxos",
				"localField":   "anchor_utxo_id",
				"foreignField": "_id",
				"as":           "result",
			},
		}},
		bson.D{{
			Key: "$unwind",
			Value: bson.M{
				"path": "$result",
			},
		}},
	}

	unspentUtxos := make([]*assetoutpoint.UnspentOutpoint, 0)
	if err := r.FindAggregate(ctx, pipeline, &unspentUtxos); err != nil {
		return nil, err
	}

	return unspentUtxos, nil
}

func NewRepoMongo(
	db *mongo.Database,
) assetoutpoint.RepoInterface {
	return &RepoMongo{
		cmrepo.NewRepoMongo(db, "asset_outpoints"),
	}
}
