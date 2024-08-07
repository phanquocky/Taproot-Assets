package onchain

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/wire"
	"go.uber.org/zap/buffer"
)

// SignRawTx function sign input btc
func (c *Client) SignRawTx(rawTx *wire.MsgTx) (*wire.MsgTx, error) {

	finalTx, isSign, err := c.client.SignRawTransaction(rawTx)
	if err != nil {
		log.Printf("cannot sign raw transaction (isSign = %v, err = %v) \n", isSign, err)
		return nil, fmt.Errorf("cannot sign raw transaction")
	}

	var signedTx buffer.Buffer

	finalTx.Serialize(&signedTx)

	log.Printf("signedTx: %x \n", signedTx)

	return finalTx, nil
}
