package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quocky/taproot-asset/server/internal/core/api"
	"github.com/quocky/taproot-asset/server/internal/domain/mint"
	mint2 "github.com/quocky/taproot-asset/taproot/model/mint"
)

// MintController define genesis controller.
type MintController struct {
	useCase mint.UseCaseInterface
}

func (c *MintController) RegisterRoutes(route gin.IRoutes) {
	route.POST("/genesis-asset", c.MintAsset)
}

func (c *MintController) MintAsset(g *gin.Context) {
	var req mint2.MintAssetReq
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, nil)

		return
	}

	_ = c.useCase.MintAsset(g, req.AmountSats, req.TapScriptRootHash, req.MintProof)

	g.JSON(http.StatusOK, nil)
}

func NewMintController(
	useCase mint.UseCaseInterface,
) api.ControllerInterface {
	return &MintController{
		useCase: useCase,
	}
}
