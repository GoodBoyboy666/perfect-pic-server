package handler

import (
	"net/http"
	"perfect-pic-server/internal/common/httpx"
	moduledto "perfect-pic-server/internal/dto"

	"github.com/gin-gonic/gin"
)

func (h *SystemHandler) GetInitState(c *gin.Context) {
	if h.initService.IsSystemInitialized() {
		c.JSON(http.StatusOK, gin.H{
			"initialized": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"initialized": false,
		})
	}
}

func (h *SystemHandler) Init(c *gin.Context) {
	var initInfo moduledto.InitRequest
	if err := c.ShouldBindJSON(&initInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}
	if h.initService.IsSystemInitialized() {
		c.JSON(http.StatusForbidden, gin.H{"error": "已初始化，无法重复初始化"})
		return
	}

	if err := h.initService.InitializeSystem(initInfo); err != nil {
		httpx.WriteServiceError(c, err, "初始化失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "初始化成功",
	})
}
