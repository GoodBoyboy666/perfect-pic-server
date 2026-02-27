package handler

import (
	"net/http"
	"perfect-pic-server/internal/common/httpx"
	moduledto "perfect-pic-server/internal/dto"

	"github.com/gin-gonic/gin"
)

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.AdminListSettings()
	if err != nil {
		httpx.WriteServiceError(c, err, "获取配置失败")
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var reqs []moduledto.UpdateSettingRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	items := make([]moduledto.UpdateSettingRequest, 0, len(reqs))
	for _, item := range reqs {
		items = append(items, moduledto.UpdateSettingRequest{Key: item.Key, Value: item.Value})
	}

	err := h.settingsService.AdminUpdateSettings(items)
	if err != nil {
		httpx.WriteServiceError(c, err, "更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "配置更新成功",
		"count":   len(reqs),
	})
}

func (h *SettingsHandler) SendTestEmail(c *gin.Context) {
	var req moduledto.SendTestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	if err := h.settingsService.AdminSendTestEmail(req.ToEmail); err != nil {
		httpx.WriteServiceError(c, err, "发送失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "测试邮件已发送"})
}
