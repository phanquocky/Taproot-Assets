package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/quocky/taproot-asset/bitcoin_runtime"
	"github.com/quocky/taproot-asset/taproot"
	"github.com/quocky/taproot-asset/taproot/address"
	"github.com/quocky/taproot-asset/taproot/config"
	"github.com/quocky/taproot-asset/taproot/onchain"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  ``,
}

var taprootClient taproot.Interface

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func newTaprootClient() taproot.Interface {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	networkCfg := config.LoadNetworkConfig()

	btcClient, err := onchain.New(networkCfg)
	if err != nil {
		log.Fatalf("Error create btc client, err: %s \n", err.Error())
	}

	wif, err := btcClient.DumpWIF(bitcoin_runtime.WalletPassphrase)
	if err != nil {
		log.Fatalf("Error dump wif, err: %s \n", err.Error())
	}

	addressMaker := address.New(networkCfg.ParamsObject)

	taprootClient := taproot.NewTaproot(btcClient, wif, addressMaker)
	log.Println("Create taproot client success!")

	return taprootClient
}

func init() {
}
