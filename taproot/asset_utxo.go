package taproot

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	utxoasset "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
	"net/http"
	"os"
	"strconv"
)

func (t *Taproot) GetAssetUTXOs(ctx context.Context, assetID string, amount int32) (*utxoasset.UnspentAssetResp, error) {
	resp, err := t.httpClient.R().
		SetContext(ctx).SetPathParams(map[string]string{
		"AssetID": assetID,
		"Amount":  strconv.FormatInt(int64(amount), 10),
		"PubKey":  hex.EncodeToString(t.wif.PrivKey.PubKey().SerializeCompressed()),
	}).
		Get(os.Getenv("SERVER_BASE_URL") + "/asset-utxo")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New("get asset UTXOs failed")
	}

	var UTXOs *utxoasset.UnspentAssetResp
	err = json.Unmarshal(resp.Body(), &UTXOs)
	if err != nil {
		return nil, err
	}

	return UTXOs, nil
}
