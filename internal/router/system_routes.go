package router

import (
	"perfect-pic-server/internal/middleware"
	systemhandler "perfect-pic-server/internal/modules/system/handler"
	"perfect-pic-server/internal/platform/service"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *systemhandler.Handler, appService *service.AppService) {
	bodyLimit := middleware.BodyLimitMiddleware(appService)

	api.GET("/init", h.GetInitState)
	api.POST("/init", bodyLimit, authLimiter, h.Init)
}
