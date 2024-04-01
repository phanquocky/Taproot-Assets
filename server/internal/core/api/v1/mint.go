package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocky/taproot-asset/server/internal/core/api"
	"github.com/quocky/taproot-asset/server/internal/domain/mint"
	"github.com/quocky/taproot-asset/server/internal/domain/utxo_asset"
	mint2 "github.com/quocky/taproot-asset/taproot/http_model/mint"
	utxoassetmodel "github.com/quocky/taproot-asset/taproot/http_model/utxo_asset"
)

// MintController define genesis controller.
type MintController struct {
	mintUseCase mint.UseCaseInterface
	utxoUseCase utxoasset.UseCaseInterface
}

func (c *MintController) RegisterRoutes(route gin.IRoutes) {
	route.POST("/mint-asset", c.MintAsset)
	route.POST("/unspent-asset-id", c.UnspentAssetsByID)
}

func (c *MintController) MintAsset(g *gin.Context) {
	var req mint2.MintAssetReq
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, nil)

		return
	}

	err := c.mintUseCase.MintAsset(g, req.AmountSats, req.TapScriptRootHash, req.MintProof)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})

		return
	}

	g.JSON(http.StatusOK, nil)
}

func (c *MintController) UnspentAssetsByID(g *gin.Context) {
	var req utxoassetmodel.UnspentAssetReq
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, nil)

		return
	}

	unspentAsset, err := c.utxoUseCase.GetUnspentAssetsById(g,
		req.AssetID,
		req.Amount,
		req.PubKey,
	)
	if err != nil {
		g.JSON(http.StatusInternalServerError, nil)

		return
	}

	g.JSON(http.StatusOK, unspentAsset)
}

func NewMintController(
	mintUseCase mint.UseCaseInterface,
	utxoUseCase utxoasset.UseCaseInterface,
) api.ControllerInterface {
	return &MintController{
		mintUseCase: mintUseCase,
		utxoUseCase: utxoUseCase,
	}
}
