package onchain

import (
	"log"
	"os"

	"github.com/btcsuite/btcd/btcutil"
)

func (c *Client) OpenWallet() error {
	err := c.client.WalletPassphrase(os.Getenv("WALLET_PASSPHRASE"), 100)
	if err != nil {
		log.Println("Cannot open wallet!")
		return err
	}

	log.Println("Open wallet success!")
	return nil
}

func (c *Client) DumpWIF() (*btcutil.WIF, error) {
	defaultAddress, err := btcutil.DecodeAddress(c.networkConfig.SenderAddress, c.networkConfig.ParamsObject)
	if err != nil {
		log.Println("[New Server] cannot decode default address, ", err)
		return nil, err
	}

	wif, err := c.client.DumpPrivKey(defaultAddress)
	if err != nil {
		log.Println("[New Server] cannot dump private key, ", err)
		return nil, err
	}

	return wif, nil

}
