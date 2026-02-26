package router

import (
	systemhandler "perfect-pic-server/internal/modules/system/handler"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *systemhandler.Handler) {
	api.GET("/init", h.GetInitState)
	api.POST("/init", authLimiter, h.Init)
}
