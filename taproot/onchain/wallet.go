package onchain

import (
	"log"

	"github.com/btcsuite/btcd/btcutil"
)

func (c *Client) openWallet(pass string) error {
	err := c.client.WalletPassphrase(pass, 100)
	if err != nil {
		return err
	}

	log.Println("[OpenWallet] Open wallet success!")
	return nil
}

func (c *Client) DumpWIF(pass string) (*btcutil.WIF, error) {
	err := c.openWallet(pass)
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
