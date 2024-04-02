package onchain

import (
	"github.com/quocky/taproot-asset/taproot/address"
	"github.com/quocky/taproot-asset/taproot/model/asset"
)

// BtcOutputInfo is struct contain all information of a btc output
type BtcOutputInfo struct {
	AddrResult  *address.TapAddress
	OutputAsset []*asset.Asset
	SatAmount   int32
}

func (r *BtcOutputInfo) GetAddrResult() *address.TapAddress {
	if r == nil {
		return nil
	}

	return r.AddrResult
}

func (r *BtcOutputInfo) GetOutputAsset() []*asset.Asset {
	if r == nil {
		return nil
	}

	return r.OutputAsset
}

func NewBtcOutputInfo(addrResult *address.TapAddress, satAmount int32, outputAsset ...*asset.Asset) *BtcOutputInfo {
	return &BtcOutputInfo{
		AddrResult:  addrResult,
		OutputAsset: outputAsset,
		SatAmount:   satAmount,
	}
}

//
//func (r *BtcOutputInfo) GetAddrResult() *address.TapAddress {
//	if r == nil {
//		return nil
//	}
//
//	return r.AddrResult
//}
//
//func (r *BtcOutputInfo) GetOutputAsset() []*asset.Asset {
//	if r == nil {
//		return nil
//	}
//
//	return r.OutputAsset
//}
