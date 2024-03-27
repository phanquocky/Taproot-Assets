package cmd

import (
	"github.com/joho/godotenv"
	"github.com/quocky/taproot-asset/taproot"
	"github.com/quocky/taproot-asset/taproot/address"
	"github.com/quocky/taproot-asset/taproot/config"
	"github.com/quocky/taproot-asset/taproot/onchain"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var TaprootClient taproot.Interface

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	networkCfg := config.LoadNetworkConfig()

	btcClient, err := onchain.New(networkCfg)
	if err != nil {
		log.Fatalf("Error create btc client, err: %s \n", err.Error())
	}

	wif, err := btcClient.DumpWIF()
	if err != nil {
		log.Fatalf("Error dump wif, err: %s \n", err.Error())
	}

	addressMaker := address.New(networkCfg.ParamsObject)

	TaprootClient = taproot.NewTaproot(btcClient, wif, addressMaker)

	log.Println("Create taproot client success!")
}
