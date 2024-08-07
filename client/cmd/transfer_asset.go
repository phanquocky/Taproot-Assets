/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/spf13/cobra"
)

// transferAssetCmd represents the transferAsset command
var transferAssetCmd = &cobra.Command{
	Use:   "transfer-asset",
	Short: "transfer-asset command is used to transfer asset to another address. Usage: transfer-asset <receiver_pubkey> <asset_id> <amount>",
	Long:  `transfer-asset command is used to transfer asset to another address. Usage: transfer-asset <receiver_pubkey> <asset_id> <amount>`,
	Run: func(cmd *cobra.Command, args []string) {
		taprootClient := newTaprootClient()

		if len(args) != 3 {
			fmt.Println("Invalid number of arguments, expected 3 arguments")
			return
		}

		receiverPubKeyStr := args[0]
		assetID := args[1]
		amount, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("Error parsing amount", err)
		}

		receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
		if err != nil {
			fmt.Println("Error decode receiver public key", err)

			return
		}

		rcvByte := [33]byte(receiverPubKey)

		rcvSerializedKey := make([]asset.SerializedKey, 1)
		rcvSerializedKey[0] = rcvByte

		err = taprootClient.TransferAsset(
			rcvSerializedKey,
			assetID,
			[]int32{int32(amount)},
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
