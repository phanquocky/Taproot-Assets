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
		taprootClient = newTaprootClient()
		pubkey := taprootClient.GetPubKey()

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
