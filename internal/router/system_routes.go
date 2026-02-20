package router

import (
	"perfect-pic-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc) {
	api.GET("/init", handler.GetInitState)
	api.POST("/init", authLimiter, handler.Init)
}
