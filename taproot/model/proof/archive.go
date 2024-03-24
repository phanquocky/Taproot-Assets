package proof

import (
	"bytes"
	"crypto/sha256"
	"github.com/quocky/taproot-asset/taproot/model/asset"

	"github.com/btcsuite/btcd/wire"
)

// Locator is able to uniquely identify a Proof in the extended Taproot Asset
// Universe by a combination of the: top-level asset ID, the group key, and also
// the script key.
type Locator struct {
	// AssetID the asset ID of the Proof to fetch. This is an optional field.
	AssetID *asset.ID

	// ScriptKey specifies the script key of the asset to fetch/store. This
	// field MUST be specified.
	ScriptKey asset.SerializedKey

	// OutPoint is the outpoint of the associated asset. This field is
	// optional.
	OutPoint *wire.OutPoint
}

// Hash returns a SHA256 Hash of the bytes serialized locator.
func (l *Locator) Hash() ([32]byte, error) {
	var buf bytes.Buffer
	if l.AssetID != nil {
		buf.Write(l.AssetID[:])
	}

	buf.Write(l.ScriptKey[:])

	if l.OutPoint != nil {
		outpointBytes := []byte(l.OutPoint.String())

		buf.Write(outpointBytes)
	}

	// Hash the buffer.
	return sha256.Sum256(buf.Bytes()), nil
}
