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
	Use:   "transfer-asset",
	Short: "transfer-asset command is used to transfer asset to another address. Usage: transfer-asset <receiver_pubkey> <asset_id> <amount>",
	Long:  `transfer-asset command is used to transfer asset to another address. Usage: transfer-asset <receiver_pubkey> <asset_id> <amount>`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("transferAsset called")
		taprootClient := newTaprootClient()

		// TODO:
		// - Get pubkey
		// - List all available assets (name, amount, assetID) - API

		// 03bcbc720d1fba2172fd413e28e778ec3a6cc640629990f97428f8beb50060faf4
		receiverPubKeyStr := "02498ecf86fb261f380e469524538b9b536a9eb1daa763001a1ddaec7b71279271"
		receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
		if err != nil {
			fmt.Println("Error decode receiver public key", err)

			return
		}

		rcvByte := [33]byte(receiverPubKey)

		rcvSerializedKey := make([]asset.SerializedKey, 1)
		rcvSerializedKey[0] = rcvByte
		// rcvSerializedKey[1] = rcvByte

		err = taprootClient.TransferAsset(
			rcvSerializedKey,
			"09f0169ce6eaa73accbcf6e54199f928be2e16344816a87be33b62159c320d25",
			[]int32{3},
		)
		if err != nil {
			fmt.Println("Error transfer asset", err)

			return
		}
	},
}

func init() {
	rootCmd.AddCommand(transferAssetCmd)
}
