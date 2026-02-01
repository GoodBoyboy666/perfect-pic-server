package service

import (
	"fmt"
	"os"
	"path/filepath"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

// GetUserStorageQuota 获取用户由于配置或默认设置计算出的实际存储配额
func GetUserStorageQuota(user *model.User) int64 {
	if user.StorageQuota != nil {
		return *user.StorageQuota
	}
	// 调用同包下的 GetInt64
	quota := GetInt64(consts.ConfigDefaultStorageQuota)
	if quota == 0 {
		return 1073741824 // 兜底 1GB
	}
	return quota
}

// DeleteUserFiles 删除指定用户的所有关联文件（头像、上传的照片）
// 此函数只负责删除物理文件，不处理数据库记录的清理
func DeleteUserFiles(userID uint) error {
	cfg := config.Get()

	// 1. 删除头像目录
	// 头像存储结构: data/avatars/{userID}/filename
	avatarRoot := cfg.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	userAvatarDir := filepath.Join(avatarRoot, fmt.Sprintf("%d", userID))

	// RemoveAll 删除路径及其包含的任何子项。如果路径不存在，RemoveAll 返回 nil（无错误）。
	if err := os.RemoveAll(userAvatarDir); err != nil {
		// 记录日志或打印错误，但不中断后续操作
		fmt.Printf("Warning: Failed to delete avatar directory for user %d: %v\n", userID, err)
	}

	// 2. 查找并删除用户上传的所有图片
	var images []model.Image
	// Unscoped() 确保即使是软删除的图片也能被查出来删除文件
	if err := db.DB.Unscoped().Where("user_id = ?", userID).Find(&images).Error; err != nil {
		return fmt.Errorf("failed to retrieve user images: %w", err)
	}

	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}

	for _, img := range images {
		// 转换路径分隔符以适配当前系统 (DB中存储的是 web 格式 '/')
		localPath := filepath.FromSlash(img.Path)
		fullPath := filepath.Join(uploadRoot, localPath)

		if err := os.Remove(fullPath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("Warning: Failed to delete image file %s: %v\n", fullPath, err)
			}
		}
	}

	return nil
}
