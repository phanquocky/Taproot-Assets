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

		// 03bcbc720d1fba2172fd413e28e778ec3a6cc640629990f97428f8beb50060faf4
		receiverPubKeyStr := "02e5119de643720aa1cbd85527a0724ec025cc7f099d31814ef311ac55972b98c7"
		// receiverPubkeyStr2 := "03c24431caaf053c9a8002a74f8738cc842edc88511516baa606ef9e354aa22167"

		receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
		if err != nil {
			fmt.Println("Error decode receiver public key", err)

			return
		}

		// receoverPubKey2, err := hex.DecodeString(receiverPubkeyStr2)
		// if err != nil {
		// 	//fmt.println("Error decode receiver public key", err)

		// 	return
		// }

		rcvByte := [33]byte(receiverPubKey)
		// rcv2Byte := [33]byte(receoverPubKey2)

		rcvSerializedKey := make([]asset.SerializedKey, 1)
		rcvSerializedKey[0] = rcvByte
		// rcvSerializedKey[1] = rcv2Byte

		err = taprootClient.TransferAsset(
			rcvSerializedKey,
			"c73c727bcf39a92f66848fe2ffe0bfd2af009a105c7f391228fa2dc53986c2a9",
			[]int32{50},
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