package onchain

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"math"
)

type UnspentTXOut struct {
	Outpoint   *wire.OutPoint
	LockScript []byte
	Amount     btcutil.Amount
}

func (c *Client) ListUTXOs() ([]*UnspentTXOut, error) {

	unspents, err := c.client.ListUnspent()
	if err != nil {
		return nil, err
	}

	var (
		UTXOs = make([]*UnspentTXOut, 0)
	)
	for _, unspent := range unspents {
		satAmount := btcutil.Amount(unspent.Amount * math.Pow10(8))

		txHash, err := chainhash.NewHashFromStr(unspent.TxID)
		if err != nil {
			return nil, err
		}

		outpoint := wire.OutPoint{Hash: *txHash, Index: unspent.Vout}

		scriptPubKey, err := hex.DecodeString(unspent.ScriptPubKey)
		if err != nil {
			return nil, err
		}

		UTXOs = append(UTXOs, &UnspentTXOut{
			Outpoint:   &outpoint,
			LockScript: scriptPubKey,
			Amount:     satAmount,
		})
	}

	return UTXOs, nil
}
