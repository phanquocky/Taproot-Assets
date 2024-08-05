package cmd

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	config "github.com/quocky/taproot-asset/server/config/core"
	"github.com/quocky/taproot-asset/server/pkg/database"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/stretchr/testify/assert"
)

// go test -timeout 30m -run ^TestProofSizeSingleAsset$ github.com/quocky/taproot-asset/client/cmd
func TestProofSizeSingleAsset(t *testing.T) {
	// log.Println("******************************** setup runtime ********************************")
	taprootClient := newTaprootClient()
	// log.Printf("******************************** setup runtime success ********************************\n")

	receiverPubKeyStr := "02498ecf86fb261f380e469524538b9b536a9eb1daa763001a1ddaec7b71279271"
	receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
	if err != nil {
		fmt.Println("Error decode receiver public key", err)
		return
	}

	rcvByte := [33]byte(receiverPubKey)
	numberOfTransfer := []int{1000} // custom this line for multiple transfer
	supply := int32(1000_000_000)

	file, err := os.Create("analyze.csv")
	assert.Nil(t, err)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"number_of_transfer", "proof_size "}
	err = writer.Write(headers)
	assert.Nil(t, err)

	cfg := config.Config{
		Mongo: config.Mongo{
			ConnURI: os.Getenv("MONGO_CONN_URI"),
			DBName:  os.Getenv("MONGO_DB_NAME"),
		},
	}

	db, err := database.NewMongoDatabase(&cfg)
	assert.Nil(t, err)

	collection := db.Collection("asset_outpoints")

	for _, num := range numberOfTransfer {
		ctx := context.Background()
		assetIDs, err := taprootClient.MintAsset(ctx, []string{"TestAsset" + fmt.Sprint(num)}, []int32{supply})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(assetIDs))
		time.Sleep(100 * time.Millisecond)

		assetID := assetIDs[0]

		for i := 0; i < num; i++ {
			err := taprootClient.TransferAsset([]asset.SerializedKey{rcvByte}, assetID, []int32{1})
			assert.Nil(t, err)
			time.Sleep(100 * time.Millisecond)
		}

		result := collection.FindOne(ctx, map[string]any{"amount": supply - int32(num), "spent": false})
		var assetOutpoint AssetOutpoint
		result.Decode(&assetOutpoint)

		fileName := assetOutpoint.ProofLocator
		data, err := os.ReadFile("../../server/locator/" + hex.EncodeToString(fileName))
		assert.Nil(t, err)

		err = writer.Write([]string{fmt.Sprint(num), fmt.Sprint(len(data))})
		assert.Nil(t, err)
	}
}

type AssetOutpoint struct {
	ProofLocator []byte `json:"proof_locator"`
}
