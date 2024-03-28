package common

import "context"

type TransactionCallbackFunc func(ctx context.Context) error

type RepoInterface interface {
	InsertOne(ctx context.Context, doc any) (ID, error)
	RunTransactions(ctx context.Context, txs []TransactionCallbackFunc) error
}
