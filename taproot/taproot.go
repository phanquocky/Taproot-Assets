package taproot

import "github.com/quocky/taproot-asset/taproot/config"

type Interface interface {
	MintAsset(names []string, amounts []int32) error
}

type Taproot struct {
}

func NewTaproot(taprootCfg *config.NetworkConfig) Interface {
	return &Taproot{}
}
