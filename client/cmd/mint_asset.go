package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// mintAssetCmd represents the mintAsset command
var mintAssetCmd = &cobra.Command{
	Use:   "mint-asset",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mintAsset called")

		names := make([]string, 0)
		amounts := make([]int32, 0)

		for _, arg := range args {
			tmp := strings.Split(arg, ":")
			names = append(names, tmp[0])
			amount, err := strconv.ParseInt(tmp[1], 10, 32)
			if err != nil {
				log.Fatalln("Error parsing amount")
			}
			amounts = append(amounts, int32(amount))
		}

		fmt.Println("Taproot client: ", TaprootClient)

		ctx := context.Background()
		err := TaprootClient.MintAsset(ctx, names, amounts)
		if err != nil {
			log.Fatalln("Error minting asset, err: ", err)
		}

		fmt.Println("Asset minted successfully")

	},
}

func init() {
	rootCmd.AddCommand(mintAssetCmd)
}
