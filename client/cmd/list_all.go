package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

// listAllCmd represents the listAll command
var listAllCmd = &cobra.Command{
	Use:   "list-all",
	Short: "List all assets",
	Long:  `List all assets in the system. This command will return all assets in the system.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("******************************** setup runtime ********************************")
		taprootClient = newTaprootClient()
		log.Printf("******************************** setup runtime success ********************************\n")

		pubkey := taprootClient.GetPubKey()

		// receiverPubKeyStr := "02498ecf86fb261f380e469524538b9b536a9eb1daa763001a1ddaec7b71279271"
		// receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
		// if err != nil {
		// 	fmt.Println("Error decode receiver public key", err)

		// 	return
		// }

		assets, err := taprootClient.ListAllAssets(context.Background(), pubkey.SerializeCompressed())
		if err != nil {
			log.Fatalf("List all assets failed, err: %s\n", err.Error())
		}

		for _, asset := range assets {
			log.Printf("Asset ID: %x, Amount: %v, Name: %v", asset.AssetID, asset.Amount, asset.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(listAllCmd)
}
