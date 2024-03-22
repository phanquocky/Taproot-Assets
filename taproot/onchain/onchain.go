package onchain

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/quocky/taproot-asset/taproot/config"
	"github.com/quocky/taproot-asset/taproot/utils"
	"log"
)

type Interface interface {
	// OpenWallet function using open wallet with passphrase
	OpenWallet() error

	// DumpWIF function using get private key of wallet
	DumpWIF() (*btcutil.WIF, error)
}

// Client struct wrap rpcclient and add some function to interact with onchain
type Client struct {
	client        *rpcclient.Client
	networkConfig *config.NetworkConfig
}

func New(networkConfig *config.NetworkConfig) (Interface, error) {
	cert, err := utils.ReadCertFile()
	if err != nil {
		log.Println("[New Server] cannot read cert file, ", err)
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
		log.Println("[New Server] cannot create new server, err ", err)
		return nil, err
	}

	return &Client{
		client:        client,
		networkConfig: networkConfig,
	}, nil
}
