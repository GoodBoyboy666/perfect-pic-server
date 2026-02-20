package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	// 注册全局安全标头中间件
	r.Use(middleware.SecurityHeaders())

	api := r.Group("/api")
	// 应用请求体大小限制中间件
	api.Use(middleware.BodyLimitMiddleware())

	// 认证限流：读取配置（在多个域路由中复用同一个实例，保持行为一致）
	authLimiter := middleware.RateLimitMiddleware(consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst)

	registerPublicRoutes(api)
	registerSystemRoutes(api, authLimiter)
	registerAuthRoutes(api, authLimiter)
	registerUserRoutes(api)
	registerAdminRoutes(api)
}
