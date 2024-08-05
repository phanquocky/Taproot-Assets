package cmd

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// mintAssetCmd represents the mintAsset command
var mintAssetCmd = &cobra.Command{
	Use:   "mint-asset",
	Short: "mint-asset command is used to mint new asset. Usage: mint-asset <asset_name>:<amount> <asset_name>:<amount> ...",
	Long:  `mint-asset command is used to mint new asset. Usage: mint-asset <asset_name>:<amount> <asset_name>:<amount> ...`,
	Run: func(cmd *cobra.Command, args []string) {
		taprootClient = newTaprootClient()
		names := make([]string, 0)
		amounts := make([]int32, 0)

		for _, arg := range args {
			tmp := strings.Split(arg, ":")
			names = append(names, tmp[0])
			amount, err := strconv.ParseInt(tmp[1], 10, 32)
			if err != nil {
				log.Fatalln("mint-asset fail error parsing amount, amount must be integer, err: ", err)
			}
			amounts = append(amounts, int32(amount))
		}

		ctx := context.Background()
		_, err := taprootClient.MintAsset(ctx, names, amounts)
		if err != nil {
			log.Fatalln("Error minting asset, err: ", err)
		}

		log.Println("Asset minted successfully")
	},
}

func init() {
	rootCmd.AddCommand(mintAssetCmd)
}
