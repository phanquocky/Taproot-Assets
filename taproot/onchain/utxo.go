package onchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// UTXOResult struct contain information when get utxos for specific amount
type UTXOResult struct {
	ActualAmount btcutil.Amount
	UTXOs        []*wire.OutPoint
	TotalAmount  btcutil.Amount
	InputFetcher *txscript.MultiPrevOutFetcher
}

// GetUTXOByAmount get utxo util the balance greater or equal amount
func (c *Client) GetUTXOByAmount(amount int32) (*UTXOResult, error) {
	inputFetcher := txscript.NewMultiPrevOutFetcher(nil)

	balance, err := c.client.GetBalance("*") // * mean all accounts
	if err != nil {
		log.Println("[Server] can not get the balance of the wallet!")
		return nil, err
	}

	btcAmount := btcutil.Amount(amount)
	if balance < btcAmount {
		return nil, fmt.Errorf("you don't have enough coin to send! your balance: %d", balance)
	}

	unspents, err := c.client.ListUnspent()
	if err != nil {
		return nil, err
	}

	var (
		inputAmount   btcutil.Amount = 0
		utxoOutpoints                = make([]*wire.OutPoint, 0)
	)
	for _, value := range unspents {

		txHash, err := chainhash.NewHashFromStr(value.TxID)
		if err != nil {
			return nil, err
		}

		outpoint := wire.OutPoint{Hash: *txHash, Index: value.Vout}
		utxoOutpoints = append(utxoOutpoints, &outpoint)

		scriptPubkey, _ := hex.DecodeString(value.ScriptPubKey)
		inputFetcher.AddPrevOut(outpoint, wire.NewTxOut(int64(value.Amount*math.Pow10(8)), scriptPubkey))

		inputAmount += btcutil.Amount(value.Amount * math.Pow10(8))

		if inputAmount >= btcAmount {
			return &UTXOResult{
				ActualAmount: btcAmount,
				UTXOs:        utxoOutpoints,
				TotalAmount:  inputAmount,
				InputFetcher: inputFetcher,
			}, nil
		}
	}

	return nil, fmt.Errorf("don't have enough utxo for amount: %d", amount)
}
