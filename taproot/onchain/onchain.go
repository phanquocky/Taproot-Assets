package onchain

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/config"
	"github.com/quocky/taproot-asset/taproot/utils"
	"log"
)

type Interface interface {
	// OpenWallet function using open wallet with passphrase
	OpenWallet() error

	// DumpWIF function using get private key of wallet
	DumpWIF() (*btcutil.WIF, error)

	// GetUTXOAmount get utxo util the balance greater or equal amount
	GetUTXOByAmount(amount int32) (*UTXOResult, error)

	// NewTxMaker function create new TxMaker struct
	NewTxMaker(utxos *UTXOResult,
		unspentAssets []*UnspentAssetsByIdResult,
		defaultOutputAmount int32,
		receivers []*Receiver,
	) (*TxMaker, error)

	// SendRawTx function send transaction to chain
	SendRawTx(rawTx *wire.MsgTx) (*chainhash.Hash, error)

	// SignRawTx function sign input btc
	SignRawTx(rawTx *wire.MsgTx) (*wire.MsgTx, error)
}

// Client struct wrap rpcclient and add some function to interact with onchain
type Client struct {
	client        *rpcclient.Client
	networkConfig *config.NetworkConfig
}

func New(networkConfig *config.NetworkConfig) (Interface, error) {
	cert, err := utils.ReadCertFile()
	if err != nil {
		return nil, err
	}

	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         networkConfig.Host,
		Params:       networkConfig.Params,
		Endpoint:     networkConfig.Endpoint,
		User:         networkConfig.User,
		Pass:         networkConfig.Pass,
		Certificates: cert,
	}, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:        client,
		networkConfig: networkConfig,
	}, nil
}

// SendRawTx function send transaction to chain
func (c *Client) SendRawTx(rawTx *wire.MsgTx) (*chainhash.Hash, error) {
	var buff bytes.Buffer

	if err := rawTx.Serialize(&buff); err != nil {
		log.Println("rawTx.Serialize(&buff)", err)
		return nil, err
	}

	log.Printf("Raw tx: %x\n", buff.Bytes())

	txHash, err := c.client.SendRawTransaction(rawTx, true)
	if err != nil {
		log.Println("cannot Send raw transaction! ", err)
		return nil, err
	}

	return txHash, nil
}

// SignRawTx function sign input btc
func (c *Client) SignRawTx(rawTx *wire.MsgTx) (*wire.MsgTx, error) {
	finalTx, isSign, err := c.client.SignRawTransaction(rawTx)
	if err != nil {
		log.Printf("cannot sign raw transaction (isSign = %v, err = %v) \n", isSign, err)
		return nil, fmt.Errorf("cannot sign raw transaction")
	}

	return finalTx, nil
}
