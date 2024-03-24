package address

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

type TapAddrMaker interface {
	CreateTapAddrByCommitment(asset.SerializedKey, *commitment.TapCommitment) (*AddrResult, error)
}

// TapAddr is used to create Taproot Address
type TapAddr struct {
	NetWork *chaincfg.Params
}

func New(params *chaincfg.Params) *TapAddr {
	return &TapAddr{
		NetWork: params,
	}
}
