package handler

import (
	"net/http"
	"perfect-pic-server/internal/modules/common/httpx"

	"github.com/gin-gonic/gin"
)

// GetServerStats 获取服务器概览统计信息
func (h *Handler) GetServerStats(c *gin.Context) {
	stats, err := h.systemService.AdminGetServerStats()
	if err != nil {
		httpx.WriteServiceError(c, err, "统计图片数据失败")
		return
	}

	c.JSON(http.StatusOK, stats)
}
