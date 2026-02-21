package handler

import (
	"net/http"
	"perfect-pic-server/internal/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var initLock sync.Mutex

func (h *Handler) GetInitState(c *gin.Context) {
	if h.service.IsSystemInitialized() {
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

	var initInfo struct {
		Username        string `json:"username" binding:"required"`
		Password        string `json:"password" binding:"required"`
		SiteName        string `json:"site_name" binding:"required"`
		SiteDescription string `json:"site_description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&initInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}
	if h.service.IsSystemInitialized() {
		c.JSON(http.StatusForbidden, gin.H{"error": "已初始化，无法重复初始化"})
		return
	}

	if err := h.service.InitializeSystem(service.InitPayload{
		Username:        initInfo.Username,
		Password:        initInfo.Password,
		SiteName:        initInfo.SiteName,
		SiteDescription: initInfo.SiteDescription,
	}); err != nil {
		WriteServiceError(c, err, "初始化失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "初始化成功",
	})
}
