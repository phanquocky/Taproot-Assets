package common

import "context"

type TransactionCallbackFunc func(ctx context.Context) error

type RepoInterface interface {
	InsertOne(ctx context.Context, doc any) (ID, error)
	FindOneByID(ctx context.Context, id ID, dest any) error
	FindMany(ctx context.Context, filter, dest any) error
	FindAggregate(ctx context.Context, filter, dest any) error
	UpdateMany(ctx context.Context, filter, update any) error
	RunTransactions(ctx context.Context, txs []TransactionCallbackFunc) error
}
