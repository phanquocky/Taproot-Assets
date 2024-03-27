package onchain

import (
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"log"
)

// SignRawTx function sign input btc
func (c *Client) SignRawTx(rawTx *wire.MsgTx) (*wire.MsgTx, error) {
	finalTx, isSign, err := c.client.SignRawTransaction(rawTx)
	if err != nil {
		log.Printf("cannot sign raw transaction (isSign = %v, err = %v) \n", isSign, err)
		return nil, fmt.Errorf("cannot sign raw transaction")
	}

	return finalTx, nil
}
