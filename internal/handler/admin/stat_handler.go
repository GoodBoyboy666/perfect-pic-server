package admin

import (
	"net/http"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"runtime"

	"github.com/gin-gonic/gin"
)

// GetServerStats 获取服务器概览统计信息
func GetServerStats(c *gin.Context) {
	var imageCount int64
	var totalSize int64
	var userCount int64

	// 统计图片数量
	if err := db.DB.Model(&model.Image{}).Count(&imageCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计图片数据失败"})
		return
	}

	// 统计占用存储空间
	// 使用 COALESCE 处理没有记录时返回 NULL 的情况
	if err := db.DB.Model(&model.Image{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计存储空间失败"})
		return
	}

	// 统计用户数量
	if err := db.DB.Model(&model.User{}).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计用户数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_count":   imageCount,
		"storage_usage": totalSize,
		"user_count":    userCount,
		"system_info": gin.H{
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
			"go_version":    runtime.Version(),
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
		},
	})
}
