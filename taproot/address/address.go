package address

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

type TapAddrMaker interface {
	CreateTapAddr(asset.SerializedKey, *commitment.TapCommitment) (*TapAddress, error)
}

type TapAddr struct {
	NetWork *chaincfg.Params
}

// AddrResult is struct contain all information of address when it is created
type AddrResult struct {
	Address           *btcutil.AddressTaproot
	TapscriptRootHash *chainhash.Hash
	Pubkey            asset.SerializedKey
	TapCommitment     *commitment.TapCommitment
}

func New(params *chaincfg.Params) *TapAddr {
	return &TapAddr{
		NetWork: params,
	}
}
