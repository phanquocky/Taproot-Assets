package taproot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/go-resty/resty/v2"
	"github.com/quocky/taproot-asset/taproot/address"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

const (
	// mock fee for every transaction on bitcoin chain
	DEFAULT_FEE = 10000

	// amount on each output contain asset commitment
	DEFAULT_OUTPUT_AMOUNT = 50

	DEFAULT_MINTING_OUTPUT_INDEX = 0

	DEFAULT_RETURN_OUTPUT_INDEX = 0

	DEFAULT_TRANSFER_OUTPUT_INDEX = 1
)

type Interface interface {
	MintAsset(ctx context.Context, names []string, amounts []int32) ([]string, error)
	GetAssetUTXOs(ctx context.Context, assetID string, amount int32) (*utxoasset.UnspentAssetResp, error)
	TransferAsset(receiverPubKey []asset.SerializedKey, assetId string, amount []int32) error
	ListAllAssets(ctx context.Context, pubkey []byte) (utxoasset.ListAssetsResp, error)
	GetPubKey() *btcec.PublicKey
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

func (t *Taproot) GetPubKey() *btcec.PublicKey {
	return t.wif.PrivKey.PubKey()
}

func (t *Taproot) ListAllAssets(ctx context.Context, pubkey []byte) (utxoasset.ListAssetsResp, error) {
	data := utxoasset.UnspentAssetReq{
		PubKey: pubkey,
	}

	resp, err := t.httpClient.R().
		SetContext(ctx).SetBody(data).Post(os.Getenv("SERVER_BASE_URL") + "/asset")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("list all assets failed")
	}

	var assets utxoasset.ListAssetsResp
	err = json.Unmarshal(resp.Body(), &assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}
