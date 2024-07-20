package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"testing"

	config "github.com/quocky/taproot-asset/server/config/core"
	"github.com/quocky/taproot-asset/server/pkg/database"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/stretchr/testify/assert"
)

func TestProofSizeSingleAsset(t *testing.T) {
	log.Println("******************************** setup runtime ********************************")
	taprootClient := newTaprootClient()
	log.Printf("******************************** setup runtime success ********************************\n")

	receiverPubKeyStr := "02498ecf86fb261f380e469524538b9b536a9eb1daa763001a1ddaec7b71279271"
	receiverPubKey, err := hex.DecodeString(receiverPubKeyStr)
	if err != nil {
		fmt.Println("Error decode receiver public key", err)
		return
	}

	rcvByte := [33]byte(receiverPubKey)
	numberOfTransfer := 1
	supply := int32(1000000)

	ctx := context.Background()
	assetIDs, err := taprootClient.MintAsset(ctx, []string{"TestAsset"}, []int32{supply})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assetIDs))

	assetID := assetIDs[0]

	for i := 0; i < numberOfTransfer; i++ {
		err := taprootClient.TransferAsset([]asset.SerializedKey{rcvByte}, assetID, []int32{1})
		assert.Nil(t, err)
	}

	cfg := config.Config{
		Mongo: config.Mongo{
			ConnURI: os.Getenv("MONGO_CONN_URI"),
			DBName:  os.Getenv("MONGO_DB_NAME"),
		},
	}

	db, err := database.NewMongoDatabase(&cfg)
	assert.Nil(t, err)

	collection := db.Collection("asset_outpoints")

	result := collection.FindOne(ctx, map[string]any{"amount": supply - int32(numberOfTransfer), "spent": true})
	var assetOutpoint AssetOutpoint
	result.Decode(&assetOutpoint)

	fileName := assetOutpoint.ProofLocator

	data, err := os.ReadFile("../../server/locator/" + string(fileName))
	assert.Nil(t, err)

	fmt.Println("Proof size:", len(data))
}

type AssetOutpoint struct {
	ProofLocator []byte `json:"proof_locator"`
}
