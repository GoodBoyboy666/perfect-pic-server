package router

import (
	"perfect-pic-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *handler.Handler) {
	api.GET("/init", h.GetInitState)
	api.POST("/init", authLimiter, h.Init)
}
