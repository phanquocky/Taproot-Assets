/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/spf13/cobra"
)

// transferAssetCmd represents the transferAsset command
var transferAssetCmd = &cobra.Command{
	Use:   "transferAsset",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("transferAsset called")

		receiverPubKeyStr := "0253a683b42c0c066d35c573977aa932ba38eb0288d54dc950e2fa50b1a49a62c6"
		receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
		if err != nil {
			fmt.Println("Error decode receiver public key", err)

			return
		}

		rcvSerializedKey := asset.SerializedKey{}
		copy(rcvSerializedKey[:], receiverPubKey)

		err = TaprootClient.TransferAsset(
			rcvSerializedKey,
			"a883c141d1aeb178f084b6cb22bcbdb6c3e302c386d4b7896d70d6cec9a02fa3",
			10,
		)
		if err != nil {
			fmt.Println("Error transfer asset", err)

			return
		}
	},
}

func init() {
	rootCmd.AddCommand(transferAssetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transferAssetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transferAssetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
