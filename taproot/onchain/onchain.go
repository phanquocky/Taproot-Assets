package onchain

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/config"
)

type Interface interface {
	DumpWIF(pass string) (*btcutil.WIF, error)
	ListUTXOs() ([]*UnspentTXOut, error)
	NewTxMaker(UTXOs []*UnspentTXOut,
		unspentAssets []*UnspentAssetsByIdResult,
		receivers []*BtcOutputInfo,
		senderBtcAddress btcutil.Address,
		fee btcutil.Amount,
	) (*TxMaker, error)

	SendRawTx(rawTx *wire.MsgTx) (*chainhash.Hash, error)
	SignRawTx(rawTx *wire.MsgTx) (*wire.MsgTx, error)
	GetSenderAddress() (btcutil.Address, error)
}

type Client struct {
	client        *rpcclient.Client
	networkConfig *config.NetworkConfig
}

func New(networkConfig *config.NetworkConfig) (Interface, error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:       networkConfig.Host,
		Params:     networkConfig.Params,
		Endpoint:   networkConfig.Endpoint,
		User:       networkConfig.User,
		Pass:       networkConfig.Pass,
		DisableTLS: true,
	}, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:        client,
		networkConfig: networkConfig,
	}, nil
}

func (c *Client) GetSenderAddress() (btcutil.Address, error) {
	return btcutil.DecodeAddress(c.networkConfig.SenderAddress, c.networkConfig.ParamsObject)
}
