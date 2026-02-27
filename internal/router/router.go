package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

type Router struct {
	authHandler     *handler.AuthHandler
	systemHandler   *handler.SystemHandler
	settingsHandler *handler.SettingsHandler
	userHandler     *handler.UserHandler
	imageHandler    *handler.ImageHandler
	service         *service.Service
}

func NewRouter(
	authHandler *handler.AuthHandler,
	systemHandler *handler.SystemHandler,
	settingsHandler *handler.SettingsHandler,
	userHandler *handler.UserHandler,
	imageHandler *handler.ImageHandler,
	appService *service.Service,
) *Router {
	return &Router{
		authHandler:     authHandler,
		systemHandler:   systemHandler,
		settingsHandler: settingsHandler,
		userHandler:     userHandler,
		imageHandler:    imageHandler,
		service:         appService,
	}
}

func (rt *Router) Init(r *gin.Engine) {
	// 注册全局安全标头中间件
	r.Use(middleware.SecurityHeaders(rt.service))

	api := r.Group("/api")

	// 认证限流：读取配置（在多个域路由中复用同一个实例，保持行为一致）
	authLimiter := middleware.RateLimitMiddleware(rt.service, consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst)

	registerPublicRoutes(api, rt.settingsHandler)
	registerSystemRoutes(api, authLimiter, rt.systemHandler, rt.service)
	registerAuthRoutes(api, authLimiter, rt.authHandler, rt.service)
	registerUserRoutes(api, rt.userHandler, rt.imageHandler, rt.service)
	registerAdminRoutes(api, rt.systemHandler, rt.settingsHandler, rt.userHandler, rt.imageHandler, rt.service)
}
