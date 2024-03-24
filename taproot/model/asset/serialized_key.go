package asset

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
)

// SerializedKey compress public key to 33 bytes
type SerializedKey [33]byte

func (s SerializedKey) ToPubKey() (*btcec.PublicKey, error) {
	return btcec.ParsePubKey(s[:])
}

func (s SerializedKey) SchnorrSerialized() []byte {
	return s[1:]
}

func (s SerializedKey) CopyBytes() []byte {
	c := make([]byte, 33)
	copy(c, s[:])

	return c
}

func ToSerialized(pubKey *btcec.PublicKey) SerializedKey {
	var serialized SerializedKey
	copy(serialized[:], pubKey.SerializeCompressed())

	return serialized
}

func StringToSerializedKey(pubkeyStr string) (SerializedKey, error) {
	pubkeyBytes, err := hex.DecodeString(pubkeyStr)

	if err != nil {
		return SerializedKey{}, err
	}

	if len(pubkeyBytes) != 33 {
		return SerializedKey{}, fmt.Errorf("pubkey is invalid len(pubkeyBytes) = %d not equal 33", len(pubkeyBytes))
	}

	var serialized SerializedKey
	copy(serialized[:], pubkeyBytes)

	return serialized, nil
}
