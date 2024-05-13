package genesisasset

import (
	genesisasset "github.com/quocky/taproot-asset/server/internal/domain/genesis_asset"
	cmrepo "github.com/quocky/taproot-asset/server/internal/repo/common"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type RepoMongo struct {
	*cmrepo.RepoMongo
}

func (r *RepoMongo) FindAvailableAssetsWithAmount(ctx context.Context, pubkey []byte) (utxoasset.ListAssetsResp, error) {
	assets := make(utxoasset.ListAssetsResp, 0)
	pipeline := mongo.Pipeline{
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "asset_outpoints",
				"localField":   "_id",
				"foreignField": "genesis_id",
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"spent":      false,
							"script_key": pubkey,
						},
					},
				},
				"as": "result",
			},
		}},
		bson.D{{
			Key: "$addFields",
			Value: bson.M{
				"amount": bson.M{
					"$sum": "$result.amount",
				},
			},
		}},
		bson.D{{
			Key: "$project",
			Value: bson.M{
				"asset_id":   1,
				"asset_name": 1,
				"amount":     1,
			},
		}},
	}

	err := r.FindAggregate(ctx, pipeline, &assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func NewRepoMongo(
	db *mongo.Database,
) genesisasset.RepoInterface {
	return &RepoMongo{
		cmrepo.NewRepoMongo(db, "genesis_assets"),
	}
}
