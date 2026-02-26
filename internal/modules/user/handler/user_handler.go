package handler

import (
	"log"
	"net/http"
	"perfect-pic-server/internal/modules/common/httpx"
	moduledto "perfect-pic-server/internal/modules/user/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetSelfInfo 获取当前用户信息
func (h *Handler) GetSelfInfo(c *gin.Context) {
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

	profile, err := h.userService.GetUserProfile(uid)
	if err != nil {
		httpx.WriteServiceError(c, err, "获取用户信息失败")
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateSelfUsername 修改自己的用户名
func (h *Handler) UpdateSelfUsername(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	var req moduledto.UpdateSelfUsernameRequest
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

	token, err := h.userService.UpdateUsernameAndGenerateToken(uid, req.Username, isAdmin)
	if err != nil {
		httpx.WriteServiceError(c, err, "更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户名更新成功",
		"token":   token,
	})
}

// UpdateSelfPassword 修改自己的密码
func (h *Handler) UpdateSelfPassword(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	var req moduledto.UpdateSelfPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	err := h.userService.UpdatePasswordByOldPassword(uid, req.OldPassword, req.NewPassword)
	if err != nil {
		httpx.WriteServiceError(c, err, "更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

func (h *Handler) RequestUpdateEmail(c *gin.Context) {
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

	var req moduledto.RequestUpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	err := h.userService.RequestEmailChange(uid, req.Password, req.NewEmail)
	if err != nil {
		httpx.WriteServiceError(c, err, "生成验证链接失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证邮件已发送至新邮箱，请查收并确认"})
}

func (h *Handler) UpdateSelfAvatar(c *gin.Context) {
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

	if h.imageService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	valid, ext, err := h.imageService.ValidateImageFile(file)
	if !valid {
		httpx.WriteServiceError(c, err, "头像文件校验失败")
		return
	}
	_ = ext
	uid, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	user, err := h.userService.GetUserByID(uid)
	if err != nil {
		httpx.WriteServiceError(c, err, "获取用户失败")
		return
	}

	newFilename, err := h.imageService.UpdateUserAvatar(user, file)
	if err != nil {
		log.Printf("UpdateUserAvatar error: %v", err)
		httpx.WriteServiceError(c, err, "头像更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功", "avatar": newFilename})
}

func (h *Handler) GetSelfImagesCount(c *gin.Context) {
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

	if h.imageService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	count, err := h.imageService.GetUserImageCount(uid)
	if err != nil {
		httpx.WriteServiceError(c, err, "获取图片数量失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_count": count,
	})
}

// BeginPasskeyRegistration 为当前登录用户发起 Passkey 绑定挑战。
func (h *Handler) BeginPasskeyRegistration(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	sessionID, creation, err := h.authService.BeginPasskeyRegistration(uid)
	if err != nil {
		httpx.WriteServiceError(c, err, "创建 Passkey 注册挑战失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":       sessionID,
		"creation_options": creation,
	})
}

// FinishPasskeyRegistration 完成当前用户的 Passkey 绑定流程。
func (h *Handler) FinishPasskeyRegistration(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	var req moduledto.FinishPasskeyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	if err := h.authService.FinishPasskeyRegistration(uid, req.SessionID, req.Credential); err != nil {
		httpx.WriteServiceError(c, err, "Passkey 绑定失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkey 绑定成功"})
}

// ListSelfPasskeys 获取当前用户已绑定的 Passkey 列表。
func (h *Handler) ListSelfPasskeys(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	passkeys, err := h.authService.ListUserPasskeys(uid)
	if err != nil {
		httpx.WriteServiceError(c, err, "获取 Passkey 列表失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": passkeys})
}

// DeleteSelfPasskey 删除当前用户指定 ID 的 Passkey。
func (h *Handler) DeleteSelfPasskey(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	idParam := c.Param("id")
	passkeyID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || passkeyID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 参数错误"})
		return
	}

	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	if err := h.authService.DeleteUserPasskey(uid, uint(passkeyID)); err != nil {
		httpx.WriteServiceError(c, err, "删除 Passkey 失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkey 删除成功"})
}

// UpdateSelfPasskeyName 修改当前用户指定 Passkey 的显示名称。
func (h *Handler) UpdateSelfPasskeyName(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "获取用户ID失败"})
		return
	}

	idParam := c.Param("id")
	passkeyID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || passkeyID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 参数错误"})
		return
	}

	var req moduledto.UpdatePasskeyNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	if err := h.authService.UpdateUserPasskeyName(uid, uint(passkeyID), req.Name); err != nil {
		httpx.WriteServiceError(c, err, "更新 Passkey 名称失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkey 名称更新成功"})
}
