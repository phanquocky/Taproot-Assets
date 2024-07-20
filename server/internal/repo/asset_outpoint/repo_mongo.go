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
				"from":         "managed_utxos",
				"localField":   "anchor_utxo_id",
				"foreignField": "_id",
				"as":           "res",
			},
		}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "asset_outpoints",
				"localField":   "anchor_utxo_id",
				"foreignField": "anchor_utxo_id",
				"let":          bson.M{"outer_id": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$ne": bson.A{"$_id", "$$outer_id"},
							},
						},
					},
					bson.D{{
						Key: "$lookup",
						Value: bson.M{
							"from":         "genesis_assets",
							"localField":   "genesis_id",
							"foreignField": "_id",
							"as":           "genesis",
						},
					}},
					bson.D{{
						Key: "$unwind",
						Value: bson.M{
							"path": "$genesis",
						},
					}},
				},
				"as": "related_assets",
			},
		}},
		bson.D{{
			Key: "$unwind",
			Value: bson.M{
				"path": "$res",
			},
		}},
		bson.D{{
			Key:   "$sort",
			Value: bson.M{"amount": -1},
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
