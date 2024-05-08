package taproot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
)

func (t *Taproot) GetAssetUTXOs(ctx context.Context, assetID string, amount int32) (*utxoasset.UnspentAssetResp, error) {
	resp, err := t.httpClient.R().
		SetContext(ctx).SetBody(map[string]any{
		"asset_id": assetID,
		"amount":   amount,
		"pub_key":  t.wif.PrivKey.PubKey().SerializeCompressed(),
	}).Post(os.Getenv("SERVER_BASE_URL") + "/unspent-asset-id")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("get asset UTXOs failed")
	}

	var UTXOs utxoasset.UnspentAssetResp
	err = json.Unmarshal(resp.Body(), &UTXOs)
	if err != nil {
		return nil, err
	}

	return &UTXOs, nil
}
