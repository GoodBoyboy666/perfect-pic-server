package handler

import (
	"net/http"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var initLock sync.Mutex

func GetInitState(c *gin.Context) {
	allow := service.GetBool(consts.ConfigAllowInit)
	if !allow {
		c.JSON(http.StatusOK, gin.H{
			"initialized": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"initialized": false,
		})
	}
}

func Init(c *gin.Context) {
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
	allowInit := service.GetBool(consts.ConfigAllowInit)
	if !allowInit {
		c.JSON(http.StatusForbidden, gin.H{"error": "已初始化，无法重复初始化"})
		return
	}
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(initInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "初始化失败",
		})
		return
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// 更新站点配置
		settingsToUpdate := map[string]string{
			consts.ConfigSiteName:        initInfo.SiteName,
			consts.ConfigSiteDescription: initInfo.SiteDescription,
			consts.ConfigAllowInit:       "false",
		}

		for key, value := range settingsToUpdate {
			if err := tx.Model(&model.Setting{}).Where("key = ?", key).Update("value", value).Error; err != nil {
				return err
			}
		}

		// 创建管理员用户
		newUser := model.User{
			Username: initInfo.Username,
			Password: string(passwordHashed),
			Avatar:   "",
			Admin:    true,
		}
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化失败"})
		return
	}
	service.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"message": "初始化成功",
	})
}
