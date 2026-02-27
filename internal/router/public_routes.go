package router

import (
	"perfect-pic-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(api *gin.RouterGroup, h *handler.SettingsHandler) {
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong from gin"})
	})
	api.GET("/webinfo", h.GetWebInfo)
	api.GET("/image_prefix", h.GetImagePrefix)
	api.GET("/avatar_prefix", h.GetAvatarPrefix)
	api.GET("/default_storage_quota", h.GetDefaultStorageQuota)
}
