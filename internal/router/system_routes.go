package router

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *handler.SystemHandler, dbConfig *config.DBConfig) {
	bodyLimit := middleware.BodyLimitMiddleware(dbConfig)

	api.GET("/init", h.GetInitState)
	api.POST("/init", bodyLimit, authLimiter, h.Init)
}
