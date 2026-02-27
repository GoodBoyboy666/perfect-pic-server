package router

import (
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *handler.SystemHandler, appService *service.Service) {
	bodyLimit := middleware.BodyLimitMiddleware(appService)

	api.GET("/init", h.GetInitState)
	api.POST("/init", bodyLimit, authLimiter, h.Init)
}
