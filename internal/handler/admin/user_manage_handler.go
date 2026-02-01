package admin

import (
	"fmt"
	"log"
	"net/http"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/middleware"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetUserList 获取用户列表
func GetUserList(c *gin.Context) {
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

	var users []model.User
	var total int64

	query := db.DB.Model(&model.User{})
	if showDeleted == "true" {
		query = query.Unscoped()
	}
	if keyword != "" {
		query = query.Where("username LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	sortOrder := "id desc"
	if order == "asc" {
		sortOrder = "id asc"
	}

	result := query.Offset((page - 1) * pageSize).Limit(pageSize).Order(sortOrder).Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 构造包含存储使用情况的响应结构
	type UserWithStorage struct {
		model.User
		StorageQuotaCalculated int64 `json:"storage_quota"`
	}

	var userList []UserWithStorage
	for _, u := range users {
		userList = append(userList, UserWithStorage{
			User:                   u,
			StorageQuotaCalculated: service.GetUserStorageQuota(&u),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      userList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUserDetail 获取指定用户信息
func GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUserRequest 创建用户请求结构体
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	if ok, msg := utils.ValidatePassword(req.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if ok, msg := utils.ValidateUsername(req.Username); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	var count int64
	db.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Admin:    false,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "data": user})
}

// UpdateUserRequest 修改用户信息结构体
type UpdateUserRequest struct {
	Username     *string `json:"username"`
	Password     *string `json:"password"`
	StorageQuota *int64  `json:"storage_quota"`
	Status       *int    `json:"status"`
}

// UpdateUser 修改用户信息
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	updates, errMsg := prepareUserUpdates(req)
	if errMsg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if len(updates) > 0 {
		if err := db.DB.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
			return
		}
		// 清除用户状态缓存
		middleware.ClearUserStatusCache(user.ID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func prepareUserUpdates(req UpdateUserRequest) (map[string]interface{}, string) {
	updates := make(map[string]interface{})

	if req.Username != nil && *req.Username != "" {
		if ok, msg := utils.ValidateUsername(*req.Username); !ok {
			return nil, msg
		}
		updates["username"] = *req.Username
	}

	if req.Password != nil && *req.Password != "" {
		if ok, msg := utils.ValidatePassword(*req.Password); !ok {
			return nil, msg
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		updates["password"] = string(hashedPassword)
	}

	if req.StorageQuota != nil {
		if *req.StorageQuota == -1 {
			updates["storage_quota"] = nil
		} else if *req.StorageQuota >= 0 {
			updates["storage_quota"] = *req.StorageQuota
		} else {
			return nil, "存储配额不能为负数（-1除外）"
		}
	}

	if req.Status != nil {
		if *req.Status == 1 || *req.Status == 2 {
			updates["status"] = *req.Status
		} else {
			return nil, "无效的用户状态"
		}
	}

	return updates, ""
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(c *gin.Context) {
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

	valid, ext, err := service.ValidateImageFile(file)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = ext

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	newFilename, err := service.UpdateUserAvatar(&user, file)
	if err != nil {
		log.Printf("Admin UpdateUserAvatar error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "头像更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像更新成功", "avatar": newFilename})
}

// RemoveUserAvatar 移除用户头像
func RemoveUserAvatar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if err := service.RemoveUserAvatar(&user); err != nil {
		log.Printf("Admin RemoveUserAvatar error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "头像移除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "头像已移除"})
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	hardDelete := c.DefaultQuery("hard_delete", "false")

	// 暂时没做禁止删除自身
	if hardDelete == "true" {
		// 先清理文件
		if err := service.DeleteUserFiles(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "清理用户文件失败"})
			return
		}

		err = db.DB.Transaction(func(tx *gorm.DB) error {
			var user model.User
			tx.Unscoped().First(&user, id)
			tx.Unscoped().Delete(&user)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
			return
		}
	} else {
		err = db.DB.Transaction(func(tx *gorm.DB) error {
			var user model.User
			tx.First(&user, id)
			// 修改名字，释放唯一索引占用，并标记为状态3(停用)
			newUsername := fmt.Sprintf("%s_del_%d", user.Username, time.Now().Unix())
			tx.Model(&user).Updates(map[string]interface{}{
				"username": newUsername,
				"status":   3,
			})
			tx.Delete(&user)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
			return
		}
	}

	// 清除用户状态缓存
	middleware.ClearUserStatusCache(uint(id))

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
