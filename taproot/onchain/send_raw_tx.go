package onchain

import (
	"bytes"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// SendRawTx function send transaction to chain
func (c *Client) SendRawTx(rawTx *wire.MsgTx) (*chainhash.Hash, error) {
	var buff bytes.Buffer

	if err := rawTx.Serialize(&buff); err != nil {
		log.Println("rawTx.Serialize(&buff)", err)
		return nil, err
	}

	log.Printf("Raw tx: %x\n", buff.Bytes())

	log.Println("SendRawTx", rawTx.TxHash().String())

	txHash, err := c.client.SendRawTransaction(rawTx, true)
	if err != nil {
		log.Println("cannot Send raw transaction! ", err)
		return nil, err
	}

	return txHash, nil
}
