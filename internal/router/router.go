package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	authMiddleware            *middleware.AuthMiddleware
	rateLimitMiddleware       *middleware.RateLimitMiddleware
	bodyLimitMiddleware       *middleware.BodyLimitMiddleware
	securityHeadersMiddleware *middleware.SecurityHeadersMiddleware
	authHandler               *handler.AuthHandler
	systemHandler             *handler.SystemHandler
	settingsHandler           *handler.SettingsHandler
	userHandler               *handler.UserHandler
	imageHandler              *handler.ImageHandler
}

func NewRouter(
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	bodyLimitMiddleware *middleware.BodyLimitMiddleware,
	securityHeadersMiddleware *middleware.SecurityHeadersMiddleware,
	authHandler *handler.AuthHandler,
	systemHandler *handler.SystemHandler,
	settingsHandler *handler.SettingsHandler,
	userHandler *handler.UserHandler,
	imageHandler *handler.ImageHandler,
) *Router {
	return &Router{
		authMiddleware:            authMiddleware,
		rateLimitMiddleware:       rateLimitMiddleware,
		bodyLimitMiddleware:       bodyLimitMiddleware,
		securityHeadersMiddleware: securityHeadersMiddleware,
		authHandler:               authHandler,
		systemHandler:             systemHandler,
		settingsHandler:           settingsHandler,
		userHandler:               userHandler,
		imageHandler:              imageHandler,
	}
}

func (rt *Router) Init(r *gin.Engine) {
	// 注册全局安全标头中间件
	r.Use(rt.securityHeadersMiddleware.SecurityHeaders())

	api := r.Group("/api")

	// 认证限流：读取配置（在多个域路由中复用同一个实例，保持行为一致）
	authLimiter := rt.rateLimitMiddleware.RateLimit(consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst)

	registerPublicRoutes(api, rt.systemHandler)
	registerSystemRoutes(api, authLimiter, rt.systemHandler, rt.bodyLimitMiddleware)
	registerAuthRoutes(api, authLimiter, rt.authHandler, rt.rateLimitMiddleware, rt.bodyLimitMiddleware)
	registerUserRoutes(api, rt.userHandler, rt.imageHandler, rt.authMiddleware, rt.bodyLimitMiddleware, rt.rateLimitMiddleware)
	registerAdminRoutes(api, rt.systemHandler, rt.settingsHandler, rt.userHandler, rt.imageHandler, rt.authMiddleware, rt.bodyLimitMiddleware)
}
