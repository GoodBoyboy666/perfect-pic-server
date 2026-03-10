package app

import (
	"log"
	"mime/multipart"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/model"
)

// ProcessImageUpload 处理图片上传核心业务
func (c *ImageUseCase) ProcessImageUpload(file *multipart.FileHeader, uid uint) (*model.Image, string, error) {
	usedSize, quota, err := c.resolveUserStorageQuota(uid)
	if err != nil {
		return nil, "", err
	}
	return c.imageService.ProcessImageUpload(file, uid, usedSize, quota)
}

// UpdateUserAvatar 更新用户头像
func (c *ImageUseCase) UpdateUserAvatar(user *model.User, file *multipart.FileHeader) (string, error) {
	newFilename, err := c.imageService.SaveUserAvatarFile(user.ID, file)
	if err != nil {
		return "", err
	}

	oldAvatar := user.Avatar
	if err := c.userService.UpdateAvatar(user, newFilename); err != nil {
		c.removeAvatarFile(user.ID, newFilename, "Rollback new avatar file")
		log.Printf("DB Update avatar error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 数据库更新失败")
	}

	c.removeAvatarFile(user.ID, oldAvatar, "Old avatar remove")
	return newFilename, nil
}

// RemoveUserAvatar 移除用户头像
func (c *ImageUseCase) RemoveUserAvatar(user *model.User) error {
	if user.Avatar == "" {
		return nil
	}

	oldAvatar := user.Avatar
	if err := c.userService.ClearAvatar(user); err != nil {
		log.Printf("DB Remove avatar error: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 移除头像失败")
	}

	c.removeAvatarFile(user.ID, oldAvatar, "Remove avatar file")
	return nil
}

func (c *ImageUseCase) resolveUserStorageQuota(uid uint) (int64, int64, error) {
	user, err := c.userStore.FindByID(uid)
	if err != nil {
		log.Printf("Get user error: %v\n", err)
		return 0, 0, commonpkg.NewInternalError("查询用户信息失败")
	}

	quota := c.dbConfig.GetDefaultStorageQuota()
	if user.StorageQuota != nil {
		quota = *user.StorageQuota
	}
	return user.StorageUsed, quota, nil
}

func (c *ImageUseCase) removeAvatarFile(userID uint, filename string, action string) {
	if filename == "" {
		return
	}
	if err := c.imageService.DeleteUserAvatarFile(userID, filename); err != nil {
		log.Printf("%s error: %v\n", action, err)
	}
}
