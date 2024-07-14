package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// getPubkeyCmd represents the get-pubkey command
var getPubkeyCmd = &cobra.Command{
	Use:   "get-pubkey",
	Short: "get-pubkey command is used to get public key, usage: get-pubkey",
	Long:  `get-pubkey command is used to get public key, usage: get-pubkey. Pubkey is used to receive asset`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("******************************** setup runtime ********************************")
		taprootClient = newTaprootClient()
		log.Printf("******************************** setup runtime success ********************************\n")

		pubkey := taprootClient.GetPubKey()
		log.Printf("Public key compress: %x", pubkey.SerializeCompressed())
		log.Printf("Public key uncompressed: %x", pubkey.SerializeUncompressed())
	},
}

func init() {
	rootCmd.AddCommand(getPubkeyCmd)
}
