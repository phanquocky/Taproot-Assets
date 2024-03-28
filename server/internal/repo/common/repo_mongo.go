package common

import (
	"context"
	"strings"

	"github.com/quocky/taproot-asset/server/internal/domain/common"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const (
	// ErrorCodeDuplicateIndexedKey duplicate indexed key of database.
	ErrorCodeDuplicateIndexedKey = "E11000"
)

// RepoMongo define common mongo repo with common function and implement function of common.RepoInterface.
type RepoMongo struct {
	db       *mongo.Database
	collName string
}

func (r *RepoMongo) InsertOne(ctx context.Context, doc any) (common.ID, error) {
	writeResult, err := r.Collection().InsertOne(ctx, doc)
	if err != nil {
		if strings.Contains(err.Error(), ErrorCodeDuplicateIndexedKey) {
			return "", common.ErrDatabaseDuplicateIndexedKey
		}

		logger.Errorw(
			"insert document into collection err",
			"collection", r.collName,
			"doc", doc,
			"err", err,
		)

		return "", common.ErrKeySystemInternalServer
	}

	objectID, ok := writeResult.InsertedID.(primitive.ObjectID)
	if !ok {
		logger.Errorw(
			"insert document into collection err",
			"collection", r.collName,
			"doc", doc,
			"err", "cast insert object to objectID err",
		)

		return "", err
	}

	return common.ID(objectID.Hex()), nil
}

func (r *RepoMongo) RunTransactions(ctx context.Context, txs []common.TransactionCallbackFunc) error {
	client := r.db.Client()

	session, err := client.StartSession()
	if err != nil {
		logger.Errorw(
			"start mongo transactions fail",
			"collection", r.collName,
			"err", err,
		)

		return err
	}

	wc := writeconcern.Majority()
	txnOptions := options.Transaction().SetWriteConcern(wc)

	defer session.EndSession(ctx)

	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		for _, tx := range txs {
			if errCallback := tx(sessionContext); errCallback != nil {
				return nil, errCallback
			}
		}

		return err, err
	}

	_, err = session.WithTransaction(ctx, callback, txnOptions)
	if err != nil {
		logger.Errorw(
			"run mongo transactions fail",
			"collection", r.collName,
			"err", err,
		)

		return err
	}

	return nil
}

func (r *RepoMongo) Collection() *mongo.Collection {
	return r.db.Collection(r.collName)
}

func NewRepoMongo(db *mongo.Database, collName string) *RepoMongo {
	return &RepoMongo{
		db:       db,
		collName: collName,
	}
}
