package router

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Router struct {
	authMiddleware  *middleware.AuthMiddleware
	bodyLimitMiddleware *middleware.BodyLimitMiddleware
	securityHeadersMiddleware *middleware.SecurityHeadersMiddleware
	authHandler     *handler.AuthHandler
	systemHandler   *handler.SystemHandler
	settingsHandler *handler.SettingsHandler
	userHandler     *handler.UserHandler
	imageHandler    *handler.ImageHandler
	dbConfig        *config.DBConfig
	gormDB          *gorm.DB
	redisDB         *redis.Client
}

func NewRouter(
	authMiddleware *middleware.AuthMiddleware,
	bodyLimitMiddleware *middleware.BodyLimitMiddleware,
	securityHeadersMiddleware *middleware.SecurityHeadersMiddleware,
	authHandler *handler.AuthHandler,
	systemHandler *handler.SystemHandler,
	settingsHandler *handler.SettingsHandler,
	userHandler *handler.UserHandler,
	imageHandler *handler.ImageHandler,
	dbConfig *config.DBConfig,
	gormDB *gorm.DB,
	redisDB *redis.Client,
) *Router {
	return &Router{
		authMiddleware:  authMiddleware,
		bodyLimitMiddleware: bodyLimitMiddleware,
		securityHeadersMiddleware: securityHeadersMiddleware,
		authHandler:     authHandler,
		systemHandler:   systemHandler,
		settingsHandler: settingsHandler,
		userHandler:     userHandler,
		imageHandler:    imageHandler,
		dbConfig:        dbConfig,
		gormDB:          gormDB,
		redisDB:         redisDB,
	}
}

func (rt *Router) Init(r *gin.Engine) {
	// 注册全局安全标头中间件
	r.Use(rt.securityHeadersMiddleware.SecurityHeaders())

	api := r.Group("/api")

	// 认证限流：读取配置（在多个域路由中复用同一个实例，保持行为一致）
	authLimiter := middleware.RateLimitMiddleware(rt.dbConfig, consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst, rt.redisDB)

	registerPublicRoutes(api, rt.systemHandler)
	registerSystemRoutes(api, authLimiter, rt.systemHandler, rt.bodyLimitMiddleware)
	registerAuthRoutes(api, authLimiter, rt.authHandler, rt.dbConfig, rt.redisDB,rt.bodyLimitMiddleware)
	registerUserRoutes(api, rt.userHandler, rt.imageHandler, rt.dbConfig, rt.authMiddleware,rt.bodyLimitMiddleware, rt.redisDB,)
	registerAdminRoutes(api, rt.systemHandler, rt.settingsHandler, rt.userHandler, rt.imageHandler, rt.authMiddleware,rt.bodyLimitMiddleware, rt.dbConfig)
}
