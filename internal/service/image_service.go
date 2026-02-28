package service

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
	"strings"

	"gorm.io/gorm"
)

// ValidateImageFile 验证上传的图片文件（大小、后缀、内容）
// 返回:
//   - bool: 是否合法
//   - string: 文件扩展名 (小写, 如 .jpg)
//   - error: 错误信息或原因
func (s *ImageService) ValidateImageFile(file *multipart.FileHeader) (bool, string, error) {
	// 检查文件大小
	maxSizeMB := s.dbConfig.GetInt(consts.ConfigMaxUploadSize) // 默认 10MB
	if file.Size > int64(maxSizeMB*1024*1024) {
		return false, "", commonpkg.NewValidationError(fmt.Sprintf("文件大小不能超过 %dMB", maxSizeMB))
	}

	// 检查文件扩展名
	allowExtsStr := s.dbConfig.GetString(consts.ConfigAllowFileExtensions)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return false, "", commonpkg.NewValidationError("无法识别文件类型")
	}

	allowed := false
	for _, allowExt := range strings.Split(allowExtsStr, ",") {
		if strings.TrimSpace(strings.ToLower(allowExt)) == ext {
			allowed = true
			break
		}
	}
	if !allowed {
		return false, ext, commonpkg.NewValidationError(fmt.Sprintf("不支持的文件类型: %s", ext))
	}

	// 检查文件内容 (Magic Bytes)
	src, err := file.Open()
	if err != nil {
		return false, ext, commonpkg.NewInternalError("无法打开上传的文件")
	}
	defer func() { _ = src.Close() }()

	if valid, msg := utils.ValidateImageContent(src, ext); !valid {
		return false, ext, commonpkg.NewValidationError(msg)
	}

	return true, ext, nil
}

// DeleteImage 删除图片文件和数据库记录
func (s *ImageService) DeleteImage(image *model.Image) error {
	cfg := config.Get()
	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 上传目录解析失败")
	}
	// 删除前先校验上传根目录节点本身，避免根目录被替换为符号链接。
	if err := utils.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		log.Printf("DeleteImage upload root security check failed: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}

	// 拼接完整物理路径
	fullPath, err := utils.SecureJoin(uploadRootAbs, image.Path)
	if err != nil {
		log.Printf("DeleteImage secure path error: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 非法文件路径")
	}

	// 使用事务确保数据库操作原子性
	if err := s.imageStore.DeleteAndDecreaseUserStorage(image); err != nil {
		return err
	}

	// 事务提交后，删除物理文件
	if err := os.Remove(fullPath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Delete file error: %v, path: %s\n", err, fullPath)
		}
	}

	return nil
}

// BatchDeleteImages 批量删除图片
func (s *ImageService) BatchDeleteImages(images []model.Image) error {
	if len(images) == 0 {
		return nil
	}

	// 使用 map 按用户分组统计待释放的空间
	// key: UserID, value: TotalSizeToFree
	userSizeMap := make(map[uint]int64)
	var imageIDs []uint
	var pathsToDelete []string

	cfg := config.Get()
	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 上传目录解析失败")
	}
	// 批量删除前先校验上传根目录节点本身，避免根目录被替换为符号链接。
	if err := utils.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		log.Printf("BatchDeleteImages upload root security check failed: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}

	for _, img := range images {
		userSizeMap[img.UserID] += img.Size
		imageIDs = append(imageIDs, img.ID)
		fullPath, secureErr := utils.SecureJoin(uploadRootAbs, img.Path)
		if secureErr != nil {
			log.Printf("BatchDeleteImages secure path error: %v\n", secureErr)
			continue
		}
		pathsToDelete = append(pathsToDelete, fullPath)
	}

	// 开启单一事务处理所有数据库变更
	if err := s.imageStore.BatchDeleteAndDecreaseUserStorage(imageIDs, userSizeMap); err != nil {
		return err
	}

	// 事务成功提交后，清理物理文件
	for _, path := range pathsToDelete {
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Batch delete file error: %v, path: %s\n", err, path)
			}
		}
	}

	return nil
}

// ListUserImages 分页查询用户自己的图片列表。
func (s *ImageService) ListUserImages(params moduledto.UserImageListRequest) ([]model.Image, int64, int, int, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)

	images, total, err := s.imageStore.ListUserImages(
		params.UserID,
		params.Filename,
		params.ID,
		(page-1)*pageSize,
		pageSize,
	)
	if err != nil {
		return nil, 0, page, pageSize, commonpkg.NewInternalError("获取图片列表失败")
	}

	return images, total, page, pageSize, nil
}

// GetUserImageCount 获取用户图片总数。
func (s *ImageService) GetUserImageCount(userID uint) (int64, error) {
	count, err := s.imageStore.CountByUserID(userID)
	if err != nil {
		return 0, commonpkg.NewInternalError("获取图片数量失败")
	}
	return count, nil
}

// GetUserOwnedImage 获取用户名下的指定图片，用于鉴权后的删除/查看。
func (s *ImageService) GetUserOwnedImage(imageID uint, userID uint) (*model.Image, error) {
	image, err := s.imageStore.FindByIDAndUserID(imageID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, commonpkg.NewNotFoundError("图片不存在或无权删除")
		}
		return nil, commonpkg.NewInternalError("查找图片失败")
	}
	return image, nil
}

// GetImagesByIDsForUser 按 ID 列表获取用户名下图片（批量场景）。
func (s *ImageService) GetImagesByIDsForUser(ids []uint, userID uint) ([]model.Image, error) {
	images, err := s.imageStore.FindByIDsAndUserID(ids, userID)
	if err != nil {
		return nil, commonpkg.NewInternalError("查找图片失败")
	}
	return images, nil
}

// DeleteUserFiles 删除指定用户的所有关联文件（头像、上传的照片）
// 此函数只负责删除物理文件，不处理数据库记录的清理
//
//nolint:gocyclo
func (s *ImageService) DeleteUserFiles(userID uint) error {
	cfg := config.Get()

	// 1. 删除头像目录
	// 头像存储结构: data/avatars/{userID}/filename
	avatarRoot := cfg.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve avatar root: %w", err)
	}
	// 先校验头像根目录节点本身，避免根目录直接是符号链接。
	if err := utils.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		return fmt.Errorf("avatar root symlink risk: %w", err)
	}

	userAvatarDir, err := utils.SecureJoin(avatarRootAbs, fmt.Sprintf("%d", userID))
	if err != nil {
		return fmt.Errorf("failed to build avatar dir: %w", err)
	}
	// 在执行 RemoveAll 前再做一次链路检查，确保目标目录链路未被并发替换为符号链接。
	if err := utils.EnsureNoSymlinkBetween(avatarRootAbs, userAvatarDir); err != nil {
		return fmt.Errorf("avatar dir symlink risk: %w", err)
	}

	// RemoveAll 删除路径及其包含的任何子项。如果路径不存在，RemoveAll 返回 nil（无错误）。
	if err := os.RemoveAll(userAvatarDir); err != nil {
		// 记录日志或打印错误，但不中断后续操作
		log.Printf("Warning: Failed to delete avatar directory for user %d: %v\n", userID, err)
	}

	// 2. 查找并删除用户上传的所有图片
	// Unscoped() 确保即使是软删除的图片也能被查出来删除文件
	images, err := s.imageStore.FindUnscopedByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user images: %w", err)
	}

	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve upload root: %w", err)
	}
	// 先校验上传根目录节点本身，避免根目录直接是符号链接。
	if err := utils.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		return fmt.Errorf("upload root symlink risk: %w", err)
	}

	for _, img := range images {
		// 转换路径分隔符以适配当前系统 (DB中存储的是 web 格式 '/')
		localPath := filepath.FromSlash(img.Path)
		fullPath, secureErr := utils.SecureJoin(uploadRootAbs, localPath)
		if secureErr != nil {
			log.Printf("Warning: Skip unsafe image path for user %d (%s): %v\n", userID, img.Path, secureErr)
			continue
		}

		if err := os.Remove(fullPath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: Failed to delete image file %s: %v\n", fullPath, err)
			}
		}
	}

	return nil
}