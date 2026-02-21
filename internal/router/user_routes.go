package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(api *gin.RouterGroup, h *handler.Handler, appService *service.AppService) {
	userGroup := api.Group("/user")
	userGroup.Use(middleware.JWTAuth())
	userGroup.Use(middleware.UserStatusCheck())

	// 修改用户名请求间隔：读取配置（秒）
	usernameLimiter := middleware.IntervalRateMiddleware(appService, consts.ConfigRateLimitUsernameUpdateIntervalSeconds)
	// 修改邮箱请求间隔：读取配置（秒）
	emailLimiter := middleware.IntervalRateMiddleware(appService, consts.ConfigRateLimitEmailUpdateIntervalSeconds)
	// 上传限流：读取配置
	uploadLimiter := middleware.RateLimitMiddleware(appService, consts.ConfigRateLimitUploadRPS, consts.ConfigRateLimitUploadBurst)
	uploadBodyLimit := middleware.UploadBodyLimitMiddleware(appService)

	userGroup.GET("/profile", h.GetSelfInfo)
	userGroup.GET("/passkeys", h.ListSelfPasskeys)
	userGroup.DELETE("/passkeys/:id", h.DeleteSelfPasskey)
	userGroup.POST("/passkeys/register/start", h.BeginPasskeyRegistration)
	userGroup.POST("/passkeys/register/finish", h.FinishPasskeyRegistration)
	userGroup.PATCH("/username", usernameLimiter, h.UpdateSelfUsername)
	userGroup.PATCH("/password", h.UpdateSelfPassword)
	userGroup.POST("/email", emailLimiter, h.RequestUpdateEmail)

	userGroup.PATCH("/avatar", uploadBodyLimit, uploadLimiter, h.UpdateSelfAvatar)

	userGroup.POST("/upload", uploadBodyLimit, uploadLimiter, h.UploadImage)
	userGroup.GET("/images", h.GetMyImages)
	userGroup.DELETE("/images/batch", h.BatchDeleteMyImages)
	userGroup.DELETE("/images/:id", h.DeleteMyImage)
	userGroup.GET("/images/count", h.GetSelfImagesCount)

	userGroup.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong with auth"})
	})
}
