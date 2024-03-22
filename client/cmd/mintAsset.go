/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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

		names := []string{}
		amounts := []int32{}

		for _, arg := range args {
			tmp := strings.Split(arg, ":")
			names = append(names, tmp[0])
			amount, err := strconv.ParseInt(tmp[1], 10, 32)
			if err != nil {
				fmt.Println("Error parsing amount")
				return
			}
			amounts = append(amounts, int32(amount))
		}

	},
}

func init() {
	rootCmd.AddCommand(mintAssetCmd)
}
