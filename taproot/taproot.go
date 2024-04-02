package taproot

import (
	"context"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/go-resty/resty/v2"
	"github.com/quocky/taproot-asset/taproot/address"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

const (
	// mock fee for every transaction on bitcoin chain
	DEFAULT_FEE = 1000

	// amount on each output contain asset commitment
	DEFAULT_OUTPUT_AMOUNT = 50

	DEFAULT_MINTING_OUTPUT_INDEX = 0

	DEFAULT_RETURN_OUTPUT_INDEX = 0

	DEFAULT_TRANSFER_OUTPUT_INDEX = 1
)

type Interface interface {
	MintAsset(ctx context.Context, names []string, amounts []int32) error
	GetAssetUTXOs(ctx context.Context, assetID string, amount int32) (*utxoasset.UnspentAssetResp, error)
	TransferAsset(receiverPubKey asset.SerializedKey, assetId string, amount int32) error
}

type Taproot struct {
	btcClient    onchain.Interface
	wif          *btcutil.WIF
	addressMaker address.TapAddrMaker
	httpClient   *resty.Client
}

func NewTaproot(btcClient onchain.Interface, wif *btcutil.WIF, addressMaker address.TapAddrMaker) Interface {
	return &Taproot{
		btcClient:    btcClient,
		wif:          wif,
		addressMaker: addressMaker,
		httpClient:   resty.New(),
	}
}
