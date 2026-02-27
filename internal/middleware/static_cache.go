package middleware

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

// StaticCacheMiddleware 为静态资源添加 Cache-Control 头
// 缓存策略由 ConfigStaticCacheControl 配置决定
func StaticCacheMiddleware(appService *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cc := appService.GetString(consts.ConfigStaticCacheControl)
		if cc != "" {
			c.Header("Cache-Control", cc)
		}
		c.Next()
	}
}
