package transfer

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	assetoutpointmodel "github.com/quocky/taproot-asset/taproot/model/asset_outpoint"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
	"golang.org/x/net/context"
)

type UseCaseInterface interface {
	TransferAsset(
		ctx context.Context,
		genesisAsset *asset.GenesisAsset,
		anchorTx *wire.MsgTx,
		amtSats int32,
		btcOutputInfos []*onchain.BtcOutputInfo,
		unspentOutpoints []*assetoutpointmodel.UnspentOutpoint,
		files []*proof.File,
	) error
}
