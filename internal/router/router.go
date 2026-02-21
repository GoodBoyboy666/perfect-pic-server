package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	adminhandler "perfect-pic-server/internal/handler/admin"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

type Router struct {
	handler      *handler.Handler
	adminHandler *adminhandler.Handler
	service      *service.AppService
}

func NewRouter(handler *handler.Handler, adminHandler *adminhandler.Handler, appService *service.AppService) *Router {
	return &Router{
		handler:      handler,
		adminHandler: adminHandler,
		service:      appService,
	}
}

func (rt *Router) Init(r *gin.Engine) {
	// 注册全局安全标头中间件
	r.Use(middleware.SecurityHeaders(rt.service))

	api := r.Group("/api")
	// 应用请求体大小限制中间件
	api.Use(middleware.BodyLimitMiddleware(rt.service))

	// 认证限流：读取配置（在多个域路由中复用同一个实例，保持行为一致）
	authLimiter := middleware.RateLimitMiddleware(rt.service, consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst)

	registerPublicRoutes(api, rt.handler)
	registerSystemRoutes(api, authLimiter, rt.handler)
	registerAuthRoutes(api, authLimiter, rt.handler, rt.service)
	registerUserRoutes(api, rt.handler, rt.service)
	registerAdminRoutes(api, rt.adminHandler)
}
