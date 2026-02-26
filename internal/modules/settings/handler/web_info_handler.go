package handler

import (
	"net/http"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/modules/settings/dto"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetWebInfo(c *gin.Context) {
	// 只获取前台展示用的公共配置项
	allowKeys := []string{
		consts.ConfigSiteName,
		consts.ConfigSiteDescription,
		consts.ConfigSiteLogo,
		consts.ConfigSiteFavicon,
	}

	var response []moduledto.WebInfoResponse
	for _, key := range allowKeys {
		val := h.settingsService.GetString(key)
		response = append(response, moduledto.WebInfoResponse{
			Key:   key,
			Value: val,
		})
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetImagePrefix(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, gin.H{
		"image_prefix": cfg.Upload.URLPrefix,
	})
}

func (h *Handler) GetAvatarPrefix(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, gin.H{
		"avatar_prefix": cfg.Upload.AvatarURLPrefix,
	})
}

func (h *Handler) GetDefaultStorageQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"default_storage_quota": h.settingsService.GetSystemDefaultStorageQuota(),
	})
}
