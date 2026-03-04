package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(
	api *gin.RouterGroup,
	userHandler *handler.UserHandler,
	imageHandler *handler.ImageHandler,
	authMiddleware *middleware.AuthMiddleware,
	bodyLimitMiddleware *middleware.BodyLimitMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
) {
	userGroup := api.Group("/user")
	userGroup.Use(authMiddleware.JWTAuth())
	userGroup.Use(authMiddleware.UserStatusCheck())
	bodyLimit := bodyLimitMiddleware.BodyLimitMiddleware()

	// 修改用户名请求间隔：读取配置（秒）
	usernameLimiter := rateLimitMiddleware.IntervalRate(consts.ConfigRateLimitUsernameUpdateIntervalSeconds)
	// 修改邮箱请求间隔：读取配置（秒）
	emailLimiter := rateLimitMiddleware.IntervalRate(consts.ConfigRateLimitEmailUpdateIntervalSeconds)
	// 上传限流：读取配置
	uploadLimiter := rateLimitMiddleware.RateLimit(consts.ConfigRateLimitUploadRPS, consts.ConfigRateLimitUploadBurst)
	uploadBodyLimit := bodyLimitMiddleware.UploadBodyLimitMiddleware()

	userGroup.GET("/profile", userHandler.GetSelfInfo)
	userGroup.GET("/passkeys", userHandler.ListSelfPasskeys)
	userGroup.DELETE("/passkeys/:id", userHandler.DeleteSelfPasskey)
	userGroup.PATCH("/passkeys/:id/name", bodyLimit, userHandler.UpdateSelfPasskeyName)
	userGroup.POST("/passkeys/register/start", bodyLimit, userHandler.BeginPasskeyRegistration)
	userGroup.POST("/passkeys/register/finish", bodyLimit, userHandler.FinishPasskeyRegistration)
	userGroup.PATCH("/username", bodyLimit, usernameLimiter, userHandler.UpdateSelfUsername)
	userGroup.PATCH("/password", bodyLimit, userHandler.UpdateSelfPassword)
	userGroup.POST("/email", bodyLimit, emailLimiter, userHandler.RequestUpdateEmail)

	userGroup.PATCH("/avatar", uploadBodyLimit, uploadLimiter, userHandler.UpdateSelfAvatar)
	userGroup.POST("/upload", uploadBodyLimit, uploadLimiter, imageHandler.UploadImage)

	userGroup.GET("/images", imageHandler.GetMyImages)
	userGroup.DELETE("/images/batch", bodyLimit, imageHandler.BatchDeleteMyImages)
	userGroup.DELETE("/images/:id", imageHandler.DeleteMyImage)
	userGroup.GET("/images/count", userHandler.GetSelfImagesCount)

	userGroup.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong with auth"})
	})
}
