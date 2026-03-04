package router

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func registerAuthRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *handler.AuthHandler, dbConfig *config.DBConfig, redisDB *redis.Client,bodyLimitMiddleware *middleware.BodyLimitMiddleware) {
	bodyLimit := bodyLimitMiddleware.BodyLimitMiddleware()

	api.POST("/login", bodyLimit, authLimiter, h.Login)
	api.POST("/register", bodyLimit, authLimiter, h.Register)
	api.POST("/auth/passkey/login/start", bodyLimit, authLimiter, h.BeginPasskeyLogin)
	api.POST("/auth/passkey/login/finish", bodyLimit, authLimiter, h.FinishPasskeyLogin)

	api.POST("/auth/email-verify", bodyLimit, h.EmailVerify)
	api.POST("/auth/email-change-verify", bodyLimit, h.EmailChangeVerify)

	// 重置密码请求间隔：读取配置（秒）
	resetLimiter := middleware.IntervalRateMiddleware(dbConfig, consts.ConfigRateLimitPasswordResetIntervalSeconds, redisDB)
	api.POST("/auth/password/reset/request", bodyLimit, resetLimiter, h.RequestPasswordReset)
	api.POST("/auth/password/reset", bodyLimit, h.ResetPassword)

	api.GET("/register", h.GetRegisterState)
	api.GET("/captcha", h.GetCaptcha)
	api.GET("/captcha/image", authLimiter, h.GetCaptchaImage)
}
