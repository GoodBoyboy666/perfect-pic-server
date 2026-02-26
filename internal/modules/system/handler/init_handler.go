package handler

import (
	"net/http"
	"perfect-pic-server/internal/modules/common/httpx"
	moduledto "perfect-pic-server/internal/modules/system/dto"
	"sync"

	"github.com/gin-gonic/gin"
)

var initLock sync.Mutex

func (h *Handler) GetInitState(c *gin.Context) {
	if h.systemService.IsSystemInitialized() {
		c.JSON(http.StatusOK, gin.H{
			"initialized": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"initialized": false,
		})
	}
}

func (h *Handler) Init(c *gin.Context) {
	// 加锁防止竞态条件
	initLock.Lock()
	defer initLock.Unlock()

	var initInfo moduledto.InitRequest
	if err := c.ShouldBindJSON(&initInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}
	if h.systemService.IsSystemInitialized() {
		c.JSON(http.StatusForbidden, gin.H{"error": "已初始化，无法重复初始化"})
		return
	}

	if err := h.systemService.InitializeSystem(initInfo); err != nil {
		httpx.WriteServiceError(c, err, "初始化失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "初始化成功",
	})
}
