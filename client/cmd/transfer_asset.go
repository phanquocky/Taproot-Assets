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
		receiverPubKeyStr := "021033ba6ebbe15348172f8c54b380eac21addb22f5f15f2c6340d609ed1a49a33"
		receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
		if err != nil {
			fmt.Println("Error decode receiver public key", err)

			return
		}

		rcvByte := [33]byte(receiverPubKey)

		rcvSerializedKey := make([]asset.SerializedKey, 2)
		rcvSerializedKey[0] = rcvByte
		// rcvSerializedKey[1] = rcvByte

		err = taprootClient.TransferAsset(
			rcvSerializedKey,
			"8e5c5da9e080437c33fc8983d67ee147f972bdb9a28c5e24996ba215d3995dcf",
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
