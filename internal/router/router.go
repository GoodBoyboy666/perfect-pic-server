package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/modules"
	"perfect-pic-server/internal/platform/service"

	"github.com/gin-gonic/gin"
)

type Router struct {
	modules *modules.AppModules
	service *service.AppService
}

func NewRouter(appModules *modules.AppModules, appService *service.AppService) *Router {
	return &Router{
		modules: appModules,
		service: appService,
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

	registerPublicRoutes(api, rt.modules.Settings.Handler)
	registerSystemRoutes(api, authLimiter, rt.modules.System.Handler)
	registerAuthRoutes(api, authLimiter, rt.modules.Auth.Handler, rt.service)
	registerUserRoutes(api, rt.modules.User.Handler, rt.modules.Image.Handler, rt.service)
	registerAdminRoutes(api, rt.modules.System.Handler, rt.modules.Settings.Handler, rt.modules.User.Handler, rt.modules.Image.Handler)
}
