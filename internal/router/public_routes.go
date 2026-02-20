package router

import (
	"perfect-pic-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(api *gin.RouterGroup) {
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong from gin"})
	})
	api.GET("/webinfo", handler.GetWebInfo)
	api.GET("/image_prefix", handler.GetImagePrefix)
	api.GET("/avatar_prefix", handler.GetAvatarPrefix)
	api.GET("/default_storage_quota", handler.GetDefaultStorageQuota)
}
