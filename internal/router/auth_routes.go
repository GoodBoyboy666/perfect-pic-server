package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/middleware"
	authhandler "perfect-pic-server/internal/modules/auth/handler"
	"perfect-pic-server/internal/platform/service"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(api *gin.RouterGroup, authLimiter gin.HandlerFunc, h *authhandler.Handler, appService *service.AppService) {
	api.POST("/login", authLimiter, h.Login)
	api.POST("/register", authLimiter, h.Register)
	api.POST("/auth/passkey/login/start", authLimiter, h.BeginPasskeyLogin)
	api.POST("/auth/passkey/login/finish", authLimiter, h.FinishPasskeyLogin)

	api.POST("/auth/email-verify", h.EmailVerify)
	api.POST("/auth/email-change-verify", h.EmailChangeVerify)

	// 重置密码请求间隔：读取配置（秒）
	resetLimiter := middleware.IntervalRateMiddleware(appService, consts.ConfigRateLimitPasswordResetIntervalSeconds)
	api.POST("/auth/password/reset/request", resetLimiter, h.RequestPasswordReset)
	api.POST("/auth/password/reset", h.ResetPassword)

	api.GET("/register", h.GetRegisterState)
	api.GET("/captcha", h.GetCaptcha)
	api.GET("/captcha/image", authLimiter, h.GetCaptchaImage)
}
