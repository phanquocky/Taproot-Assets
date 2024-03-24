package config

import (
	"github.com/btcsuite/btcd/chaincfg"
	"os"
)

func LoadNetworkConfig() *NetworkConfig {
	params := "testnet3"
	paramsObject := &chaincfg.TestNet3Params
	host := "localhost:18332"

	env := os.Getenv("ENV")
	if env == "sim" {
		params = "simnet"
		paramsObject = &chaincfg.SimNetParams
		host = "localhost:18554"
	}

	return &NetworkConfig{
		Host:          host,
		Endpoint:      os.Getenv("ENDPOINT_CONFIG"),
		User:          os.Getenv("USER_CONFIG"),
		Pass:          os.Getenv("PASS_CONFIG"),
		Params:        params,
		ParamsObject:  paramsObject,
		SenderAddress: os.Getenv("SENDER_ADDR"),
	}
}
