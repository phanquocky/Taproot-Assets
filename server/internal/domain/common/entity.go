package common

import (
	"errors"
	"time"
)

var (
	ErrKeySystemInternalServer     = errors.New("error.system.internal")
	ErrDatabaseNotFound            = errors.New("error.database.not_found_data")
	ErrDatabaseDuplicateIndexedKey = errors.New("error.database.duplicate_indexed_key")
)

// ID is a custom type that helps to Marshal id value from Database.
// to string in the JSON response.
type ID string

// UnixTimestamp is a custom type that helps to Marshal a Datetime value from Database.
// to Unix epoch timestamp in the JSON response.
type UnixTimestamp time.Time

// CreatedAt is a custom type that helps to Marshal a Datetime value from Database.
// to Unix epoch timestamp in the JSON response and auto upsert time when upsert document.
type CreatedAt UnixTimestamp

type Entity struct {
	ID        ID        `json:"id" bson:"_id"`
	CreatedAt CreatedAt `json:"created_at"`
	UpdatedAt CreatedAt `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}
