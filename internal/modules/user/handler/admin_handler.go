package handler

import (
	"log"
	"net/http"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/modules/common/httpx"
	moduledto "perfect-pic-server/internal/modules/user/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserList 获取用户列表
func (h *Handler) GetUserList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	keyword := c.Query("keyword")
	showDeleted := c.DefaultQuery("show_deleted", "false")
	order := c.Query("order")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	users, total, err := h.userService.AdminListUsers(moduledto.AdminUserListRequest{
		Page:        page,
		PageSize:    pageSize,
		Keyword:     keyword,
		ShowDeleted: showDeleted == "true",
		Order:       order,
	})
	if err != nil {
		httpx.WriteServiceError(c, err, "获取用户列表失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUserDetail 获取指定用户信息
func (h *Handler) GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := h.userService.AdminGetUserDetail(uint(id))
	if err != nil {
		httpx.WriteServiceError(c, err, "获取用户失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUser 创建用户
func (h *Handler) CreateUser(c *gin.Context) {
	var req moduledto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	user, err := h.userService.AdminCreateUser(moduledto.AdminCreateUserRequest(req))
	if err != nil {
		httpx.WriteServiceError(c, err, "创建用户失败")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "data": user})
}

// UpdateUser 修改用户信息
func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req moduledto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	updates, err := h.userService.AdminPrepareUserUpdates(uint(id), moduledto.AdminUserUpdateRequest(req))
	if err != nil {
		httpx.WriteServiceError(c, err, "更新用户失败")
		return
	}

	if len(updates) > 0 {
		if err := h.userService.AdminApplyUserUpdates(uint(id), updates); err != nil {
			httpx.WriteServiceError(c, err, "更新用户失败")
			return
		}
		// 清除用户状态缓存
		middleware.ClearUserStatusCache(uint(id))
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// UpdateUserAvatar 更新用户头像
func (h *Handler) UpdateUserAvatar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
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

	user, err := h.userService.AdminGetUserDetail(uint(id))
	if err != nil {
		httpx.WriteServiceError(c, err, "获取用户失败")
		return
	}

	newFilename, err := h.imageService.UpdateUserAvatar(user, file)
	if err != nil {
		log.Printf("Admin UpdateUserAvatar error: %v", err)
		httpx.WriteServiceError(c, err, "头像更新失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功", "avatar": newFilename})
}

// RemoveUserAvatar 移除用户头像
func (h *Handler) RemoveUserAvatar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := h.userService.AdminGetUserDetail(uint(id))
	if err != nil {
		httpx.WriteServiceError(c, err, "获取用户失败")
		return
	}

	if h.imageService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务不可用"})
		return
	}

	if err := h.imageService.RemoveUserAvatar(user); err != nil {
		log.Printf("Admin RemoveUserAvatar error: %v", err)
		httpx.WriteServiceError(c, err, "头像移除失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像已移除"})
}

// DeleteUser 删除用户
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	hardDelete := c.DefaultQuery("hard_delete", "false")

	if err := h.userService.AdminDeleteUser(uint(id), hardDelete == "true"); err != nil {
		httpx.WriteServiceError(c, err, "删除用户失败")
		return
	}

	// 清除用户状态缓存
	middleware.ClearUserStatusCache(uint(id))

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
