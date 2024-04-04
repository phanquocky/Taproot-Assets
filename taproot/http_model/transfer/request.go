package transfer

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

type TransferReq struct {
	GenesisAsset     *asset.GenesisAsset                   `json:"genesis_asset"`
	AnchorTx         *wire.MsgTx                           `json:"anchor_tx"`
	AmtSats          int32                                 `json:"amt_sats"`
	BtcOutputInfos   []*onchain.BtcOutputInfo              `json:"btc_output_infos"`
	UnspentOutpoints []*assetoutpointmodel.UnspentOutpoint `json:"unspent_outpoints"`
	Files            []*proof.File                         `json:"files"`
}
