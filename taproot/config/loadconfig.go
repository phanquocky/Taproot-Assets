package config

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/quocky/taproot-asset/bitcoin_runtime"
)

func LoadNetworkConfig() *NetworkConfig {
	params := "testnet3"
	paramsObject := &chaincfg.TestNet3Params
	host := fmt.Sprintf("localhost:%s", os.Getenv("WPORT"))
	senderAddress := bitcoin_runtime.MiningAddr
	user := bitcoin_runtime.MockBtcUser
	pass := bitcoin_runtime.MockBtcPass

	env := os.Getenv("ENV")
	if env == "sim" {
		params = "simnet"
		paramsObject = &chaincfg.SimNetParams
	}

	return &NetworkConfig{
		Host:          host,
		Endpoint:      "ws",
		User:          user,
		Pass:          pass,
		Params:        params,
		ParamsObject:  paramsObject,
		SenderAddress: senderAddress,
	}
}
