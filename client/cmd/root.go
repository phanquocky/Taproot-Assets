package cmd

import (
	"github.com/joho/godotenv"
	"github.com/quocky/taproot-asset/taproot"
	"github.com/quocky/taproot-asset/taproot/config"
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
	TaprootClient, err = taproot.NewTaproot(networkCfg)
	if err != nil {
		log.Fatalf("Error create taproot client, err: %s", err.Error())
	}
}
