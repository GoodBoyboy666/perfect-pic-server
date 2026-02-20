package handler

import (
	"log"
	"net/http"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetSelfInfo 获取当前用户信息
func GetSelfInfo(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	profile, err := service.GetUserProfile(uid)
	if err != nil {
		WriteServiceError(c, err, "获取用户信息失败")
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateSelfUsername 修改自己的用户名
func UpdateSelfUsername(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 获取管理员权限状态
	adminVal, _ := c.Get("admin")
	isAdmin := false
	if val, ok := adminVal.(bool); ok {
		isAdmin = val
	}

	token, err := service.UpdateUsernameAndGenerateToken(uid, req.Username, isAdmin)
	if err != nil {
		WriteServiceError(c, err, "更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户名更新成功",
		"token":   token,
	})
}

// UpdateSelfPassword 修改自己的密码
func UpdateSelfPassword(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	err := service.UpdatePasswordByOldPassword(uid, req.OldPassword, req.NewPassword)
	if err != nil {
		WriteServiceError(c, err, "更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

func RequestUpdateEmail(c *gin.Context) {
	id, _ := c.Get("id")
	if id == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在"})
		return
	}

	uid, ok := id.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID类型错误"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
		NewEmail string `json:"new_email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	err := service.RequestEmailChange(uid, req.Password, req.NewEmail)
	if err != nil {
		WriteServiceError(c, err, "生成验证链接失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证邮件已发送至新邮箱，请查收并确认"})
}

func UpdateSelfAvatar(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	valid, ext, err := service.ValidateImageFile(file)
	if !valid {
		WriteServiceError(c, err, "头像文件校验失败")
		return
	}
	_ = ext
	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	user, err := service.GetUserByID(uid)
	if err != nil {
		WriteServiceError(c, err, "获取用户失败")
		return
	}

	newFilename, err := service.UpdateUserAvatar(user, file)
	if err != nil {
		log.Printf("UpdateUserAvatar error: %v", err)
		WriteServiceError(c, err, "头像更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功", "avatar": newFilename})
}

func GetSelfImagesCount(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	count, err := service.GetUserImageCount(uid)
	if err != nil {
		WriteServiceError(c, err, "获取图片数量失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_count": count,
	})
}
