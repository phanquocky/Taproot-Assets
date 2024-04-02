package transfer

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	manageutxo "github.com/quocky/taproot-asset/taproot/model/manage_utxo"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
)

type TransferReq struct {
	GenesisAsset     *asset.GenesisAsset
	AnchorTx         *wire.MsgTx
	AmtSats          int32
	BtcOutputInfos   []*onchain.BtcOutputInfo
	UnspentOutpoints []*manageutxo.UnspentOutpoint
	Files            []*proof.File
}
