package router

import (
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *handler.SystemHandler, bodyLimitMiddleware *middleware.BodyLimitMiddleware) {
	bodyLimit := bodyLimitMiddleware.BodyLimitMiddleware()

	api.GET("/init", h.GetInitState)
	api.POST("/init", bodyLimit, authLimiter, h.Init)
}
