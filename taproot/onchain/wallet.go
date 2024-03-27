package onchain

import (
	"log"
	"os"

	"github.com/btcsuite/btcd/btcutil"
)

func (c *Client) OpenWallet() error {
	err := c.client.WalletPassphrase(os.Getenv("WALLET_PASSPHRASE"), 100)
	if err != nil {
		return err
	}

	log.Println("[OpenWallet] Open wallet success!")
	return nil
}

func (c *Client) DumpWIF() (*btcutil.WIF, error) {
	err := c.OpenWallet()
	if err != nil {
		return nil, err
	}

	defaultAddress, err := btcutil.DecodeAddress(c.networkConfig.SenderAddress, c.networkConfig.ParamsObject)
	if err != nil {
		return nil, err
	}

	wif, err := c.client.DumpPrivKey(defaultAddress)
	if err != nil {
		return nil, err
	}

	return wif, nil

}
