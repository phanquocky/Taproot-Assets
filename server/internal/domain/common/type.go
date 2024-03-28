package common

import (
	"fmt"
	"strconv"
	"time"

	"github.com/quocky/taproot-asset/server/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// define common constant.
const (
	DecimalBase     = 10
	hexadecimal     = 16
	Hundred         = 100
	ZeroAddress     = "0x0000000000000000000000000000000000000000"
	ZeroObjectIDStr = "000000000000000000000000"
)

// String returns string of ID type.
func (id ID) String() string {
	return string(id)
}

// Hex returns hex string of ID type.
func (id ID) Hex() string {
	objectID, err := primitive.ObjectIDFromHex(id.String())
	if err != nil {
		logger.Errorw("gender hex string from id err", "id", id)
	}

	return objectID.Hex()
}

// IsEmpty check id is "".
func (id ID) IsEmpty() bool {
	return id.String() == ""
}

// NewID returns new random ID.
func NewID() ID {
	return ID(primitive.NewObjectID().Hex())
}

// GetBSON helps to store ID as primitive.ObjectID in database.
//
// specific use for mongoDB.
func (id ID) GetBSON() (any, error) {
	if id.String() == "" {
		return primitive.NewObjectID(), nil
	}

	objID, err := primitive.ObjectIDFromHex(id.String())
	if err != nil {
		return nil, err
	}

	return objID, nil
}

// SetBSON creates ID from a MongoDB primitive.ObjectID.
func (id *ID) SetBSON(raw bson.RawValue) error {
	var idVal string

	if err := raw.Unmarshal(&idVal); err != nil {
		return err
	}

	*id = ID(idVal)

	return nil
}

// Time returns time.Time value.
func (t *UnixTimestamp) Time() time.Time {
	return time.Time(*t)
}

// MarshalJSON helps to parse Datetime from database to unix epoch timestamp.
func (t UnixTimestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// UnmarshalJSON cast time.Time or unix epoch timestamp input to UnixTimestamp type.
func (t *UnixTimestamp) UnmarshalJSON(b []byte) error {
	unix, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		ti := time.Time(*t)
		res := ti.UnmarshalJSON(b)
		*t = UnixTimestamp(ti)

		return res
	}

	ti := time.Unix(unix, 0)
	*t = UnixTimestamp(ti)

	return nil
}

// GetBSON helps to store timestamp as Datetime in database.
func (t UnixTimestamp) GetBSON() (interface{}, error) {
	return time.Time(t), nil
}

// SetBSON creates a UnixTimestamp from a MongoDB datetime string.
func (t *UnixTimestamp) SetBSON(raw bson.RawValue) error {
	// create datetime type object to reuse its UnmarshalJSON method
	datetime := primitive.NewDateTimeFromTime(t.Time())

	err := raw.Unmarshal(&datetime)
	if err == nil {
		// set back unmarshalled value
		*t = UnixTimestamp(datetime.Time())

		return nil
	}

	var unix int64

	err = raw.Unmarshal(&unix)
	if err != nil {
		return err
	}

	ti := time.Unix(unix, 0)
	*t = UnixTimestamp(ti)

	return nil
}

// Time returns time.Time value.
func (c *CreatedAt) Time() time.Time {
	unixTime := UnixTimestamp(*c)

	return unixTime.Time()
}

// MarshalJSON helps to parse Datetime from database to unix epoch timestamp.
func (c CreatedAt) MarshalJSON() ([]byte, error) {
	ts := c.Time().Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// UnmarshalJSON cast time.Time or unix epoch timestamp input to UnixTimestamp type.
func (c *CreatedAt) UnmarshalJSON(b []byte) error {
	unix, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		ti := time.Time(*c)
		res := ti.UnmarshalJSON(b)
		*c = CreatedAt(ti)

		return res
	}

	ti := time.Unix(unix, 0)
	*c = CreatedAt(ti)

	return nil
}

// GetBSON will run when inserting or updating createdAt field to DB, this helps to auto set time.
func (c CreatedAt) GetBSON() (any, error) {
	if c.Time().IsZero() {
		return time.Now(), nil
	}

	return c.Time(), nil
}

// SetBSON will run to decode BSON raw value.
func (c *CreatedAt) SetBSON(raw bson.RawValue) error {
	// create datetime type object to reuse its UnmarshalJSON method
	datetime := primitive.NewDateTimeFromTime(c.Time())

	err := raw.Unmarshal(&datetime)
	if err == nil {
		// set back unmarshalled value
		*c = CreatedAt(datetime.Time())

		return nil
	}

	return nil
}
