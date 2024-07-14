package database

import (
	"context"
	"fmt"

	config "github.com/quocky/taproot-asset/server/config/core"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/mgocompat"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewMongoDatabase(cfg *config.Config) (*mongo.Database, error) {
	ctx := context.Background()

	fmt.Println("URI:", cfg.Mongo.ConnURI)

	bsonOpts := &options.BSONOptions{UseJSONStructTags: true}

	client, err := mongo.Connect(
		ctx,
		options.Client().SetRegistry(mgocompat.Registry),
		options.Client().SetReadPreference(readpref.Secondary()),
		options.Client().ApplyURI(cfg.Mongo.ConnURI),
		options.Client().SetBSONOptions(bsonOpts),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	logger.Infow("Connect to MongoDB successfully")

	db := client.Database(cfg.Mongo.DBName)

	return db, nil
}
