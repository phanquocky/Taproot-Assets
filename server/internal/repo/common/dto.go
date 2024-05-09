package common

import (
	"golang.org/x/net/context"
)

type cursor interface {
	Decode(v any) error
}

func DecodeOne(ctx context.Context, cursor cursor, dest any) error {
	return cursor.Decode(dest)
}
