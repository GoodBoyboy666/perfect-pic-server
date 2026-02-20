package router

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/handler"
	"perfect-pic-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(api *gin.RouterGroup) {
	userGroup := api.Group("/user")
	userGroup.Use(middleware.JWTAuth())
	userGroup.Use(middleware.UserStatusCheck())

	// 修改用户名请求间隔：读取配置（秒）
	usernameLimiter := middleware.IntervalRateMiddleware(consts.ConfigRateLimitUsernameUpdateIntervalSeconds)
	// 修改邮箱请求间隔：读取配置（秒）
	emailLimiter := middleware.IntervalRateMiddleware(consts.ConfigRateLimitEmailUpdateIntervalSeconds)
	// 上传限流：读取配置
	uploadLimiter := middleware.RateLimitMiddleware(consts.ConfigRateLimitUploadRPS, consts.ConfigRateLimitUploadBurst)
	uploadBodyLimit := middleware.UploadBodyLimitMiddleware()

	userGroup.GET("/profile", handler.GetSelfInfo)
	userGroup.PATCH("/username", usernameLimiter, handler.UpdateSelfUsername)
	userGroup.PATCH("/password", handler.UpdateSelfPassword)
	userGroup.POST("/email", emailLimiter, handler.RequestUpdateEmail)

	userGroup.PATCH("/avatar", uploadBodyLimit, uploadLimiter, handler.UpdateSelfAvatar)

	userGroup.POST("/upload", uploadBodyLimit, uploadLimiter, handler.UploadImage)
	userGroup.GET("/images", handler.GetMyImages)
	userGroup.DELETE("/images/batch", handler.BatchDeleteMyImages)
	userGroup.DELETE("/images/:id", handler.DeleteMyImage)
	userGroup.GET("/images/count", handler.GetSelfImagesCount)

	userGroup.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong with auth"})
	})
}
