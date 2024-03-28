package api

import "github.com/gin-gonic/gin"

// ControllerInterface define interface controller for all controller of http api.
type ControllerInterface interface {
	RegisterRoutes(route gin.IRoutes)
}

// RegisterRoutes register the routes of controller class.
func RegisterRoutes(router gin.IRoutes, controllers ...ControllerInterface) {
	for _, item := range controllers {
		item.RegisterRoutes(router)
	}
}
