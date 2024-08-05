package main

import (
	"errors"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	config "github.com/quocky/taproot-asset/server/config/core"
	"github.com/quocky/taproot-asset/server/internal/core/api"
	v1 "github.com/quocky/taproot-asset/server/internal/core/api/v1"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/repo/asset_outpoint"
	chaintx "github.com/quocky/taproot-asset/server/internal/repo/chain_tx"
	genesisasset "github.com/quocky/taproot-asset/server/internal/repo/genesis_asset"
	genesispoint "github.com/quocky/taproot-asset/server/internal/repo/genesis_point"
	manageutxo "github.com/quocky/taproot-asset/server/internal/repo/manage_utxo"
	mintU "github.com/quocky/taproot-asset/server/internal/usecase/mint"
	transferU "github.com/quocky/taproot-asset/server/internal/usecase/transfer"
	utxoU "github.com/quocky/taproot-asset/server/internal/usecase/utxo"
	"github.com/quocky/taproot-asset/server/pkg/database"
	"github.com/quocky/taproot-asset/server/pkg/logger"
	networkCfg "github.com/quocky/taproot-asset/taproot/config"
)

func main() {
	logger.Init()

	cfg := config.NewConfig()

	db, err := database.NewMongoDatabase(cfg)
	if err != nil {
		panic(err)
	}

	rpcClient, err := NewRPCClient()
	if err != nil {
		panic(err)
	}

	router := NewServer()

	// repo
	genesisAssetRepo := genesisasset.NewRepoMongo(db)
	assetOutpointRepo := assetoutpoint.NewRepoMongo(db)
	chainTxRepo := chaintx.NewRepoMongo(db)
	genesisPointRepo := genesispoint.NewRepoMongo(db)
	manageUtxoRepo := manageutxo.NewRepoMongo(db)

	// use case
	mintUseCase := mintU.NewUseCase(genesisAssetRepo, assetOutpointRepo, chainTxRepo, genesisPointRepo, manageUtxoRepo, rpcClient)
	utxoUseCase := utxoU.NewUseCase(genesisAssetRepo, assetOutpointRepo, genesisPointRepo)
	transferUseCase := transferU.NewUseCase(assetOutpointRepo, chainTxRepo, manageUtxoRepo, genesisAssetRepo, rpcClient)

	// controller
	mintController := v1.NewMintController(mintUseCase, utxoUseCase, transferUseCase)

	// register routes
	api.RegisterRoutes(router, mintController)

	router.Run()
}

func NewServer() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	return engine
}

func NewRPCClient() (*rpcclient.Client, error) {
	networkCfg := networkCfg.LoadNetworkConfig()
	networkCfg.Host = "localhost:8000" // TODO: move to config

	rpcCfg := &rpcclient.ConnConfig{
		Host:       networkCfg.Host,
		Endpoint:   networkCfg.Endpoint,
		User:       networkCfg.User,
		Pass:       networkCfg.Pass,
		Params:     networkCfg.Params,
		DisableTLS: true,
	}

	rpcClient, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, errors.New("create RPC client fail, " + err.Error())
	}

	return rpcClient, nil
}
