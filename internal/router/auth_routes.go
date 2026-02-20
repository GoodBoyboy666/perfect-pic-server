package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc) {
	api.POST("/login", authLimiter, handler.Login)
	api.POST("/register", authLimiter, handler.Register)

	api.POST("/auth/email-verify", handler.EmailVerify)
	api.POST("/auth/email-change-verify", handler.EmailChangeVerify)

	// 重置密码请求间隔：读取配置（秒）
	resetLimiter := middleware.IntervalRateMiddleware(consts.ConfigRateLimitPasswordResetIntervalSeconds)
	api.POST("/auth/password/reset/request", resetLimiter, handler.RequestPasswordReset)
	api.POST("/auth/password/reset", handler.ResetPassword)

	api.GET("/register", handler.GetRegisterState)
	api.GET("/captcha", handler.GetCaptcha)
	api.GET("/captcha/image", authLimiter, handler.GetCaptchaImage)
}
