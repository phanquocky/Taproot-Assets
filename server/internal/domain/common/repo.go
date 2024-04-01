package common

import "context"

type TransactionCallbackFunc func(ctx context.Context) error

type RepoInterface interface {
	InsertOne(ctx context.Context, doc any) (ID, error)
	FindOneByID(ctx context.Context, id ID, dest any) error
	FindMany(ctx context.Context, filter any, dest any) error
	FindAggregate(ctx context.Context, filter any, dest any) error
	RunTransactions(ctx context.Context, txs []TransactionCallbackFunc) error
}
