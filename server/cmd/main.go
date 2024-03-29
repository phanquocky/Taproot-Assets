package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/quocky/taproot-asset/server/config/core"
	"github.com/quocky/taproot-asset/server/internal/core/api"
	v1 "github.com/quocky/taproot-asset/server/internal/core/api/v1"
	"github.com/quocky/taproot-asset/server/internal/repo/asset"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/repo/asset_outpoint"
	chaintx "github.com/quocky/taproot-asset/server/internal/repo/chain_tx"
	genesispoint "github.com/quocky/taproot-asset/server/internal/repo/genesis_point"
	manageutxo "github.com/quocky/taproot-asset/server/internal/repo/manage_utxo"
	mintU "github.com/quocky/taproot-asset/server/internal/usecase/mint"
	"github.com/quocky/taproot-asset/server/pkg/database"
	"github.com/quocky/taproot-asset/server/pkg/logger"
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
	assetRepo := asset.NewRepoMongo(db)
	assetOutpointRepo := assetoutpoint.NewRepoMongo(db)
	chainTxRepo := chaintx.NewRepoMongo(db)
	genesisPointRepo := genesispoint.NewRepoMongo(db)
	manageUtxoRepo := manageutxo.NewRepoMongo(db)

	// use case
	mintUseCase := mintU.NewUseCase(assetRepo, assetOutpointRepo, chainTxRepo, genesisPointRepo, manageUtxoRepo, rpcClient)

	// controller
	mintController := v1.NewMintController(mintUseCase)

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
	certPath := filepath.Join(btcutil.AppDataDir("btcd", false), "rpc.cert")

	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, errors.New("cannot read cert file, " + err.Error())
	}

	rpcCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv("IP"),
		Endpoint:     "ws",
		User:         os.Getenv("USER_CONFIG"),
		Pass:         os.Getenv("PASS_CONFIG"),
		Certificates: cert,
	}

	rpcClient, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, errors.New("create RPC client fail, " + err.Error())
	}

	return rpcClient, nil
}
