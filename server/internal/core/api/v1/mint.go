package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocky/taproot-asset/server/internal/core/api"
	"github.com/quocky/taproot-asset/server/internal/domain/mint"
	mint2 "github.com/quocky/taproot-asset/taproot/http_model/mint"
)

// MintController define genesis controller.
type MintController struct {
	useCase mint.UseCaseInterface
}

func (c *MintController) RegisterRoutes(route gin.IRoutes) {
	route.POST("/mint-asset", c.MintAsset)
}

func (c *MintController) MintAsset(g *gin.Context) {
	var req mint2.MintAssetReq
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, nil)

		return
	}

	err := c.useCase.MintAsset(g, req.AmountSats, req.TapScriptRootHash, req.MintProof)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})

		return
	}

	g.JSON(http.StatusOK, nil)
}

func NewMintController(
	useCase mint.UseCaseInterface,
) api.ControllerInterface {
	return &MintController{
		useCase: useCase,
	}
}
