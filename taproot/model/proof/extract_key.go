package proof

import (
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// ExtractTaprootKey attempts to extract a Taproot tweaked key from the output
// found at `outputIndex`.
func ExtractTaprootKey(tx *wire.MsgTx,
	outputIndex uint32) (*btcec.PublicKey, error) {

	if outputIndex >= uint32(len(tx.TxOut)) {
		return nil, errors.New("invalid output index")
	}

	return ExtractTaprootKeyFromScript(tx.TxOut[outputIndex].PkScript)
}

// ExtractTaprootKeyFromScript attempts to extract a Taproot tweaked key from
// the given output script.
func ExtractTaprootKeyFromScript(pkScript []byte) (*btcec.PublicKey, error) {
	version, keyBytes, err := txscript.ExtractWitnessProgramInfo(pkScript)
	if err != nil {
		return nil, err
	}

	if version != txscript.TaprootWitnessVersion {
		return nil, errors.New("invalid witness version")
	}

	return schnorr.ParsePubKey(keyBytes)
}
