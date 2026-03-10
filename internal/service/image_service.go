package service

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/pkg/pathpkg"
	"perfect-pic-server/internal/pkg/validator"
	repo "perfect-pic-server/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
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

	if valid, msg := validator.ValidateImageContent(src, ext); !valid {
		return false, ext, commonpkg.NewValidationError(msg)
	}

	return true, ext, nil
}

// DeleteImage 删除图片文件和数据库记录
func (s *ImageService) DeleteImage(image *model.Image) error {
	cfg := s.staticConfig
	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 上传目录解析失败")
	}
	// 删除前先校验上传根目录节点本身，避免根目录被替换为符号链接。
	if err := pathpkg.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		log.Printf("DeleteImage upload root security check failed: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}

	// 拼接完整物理路径
	fullPath, err := pathpkg.SecureJoin(uploadRootAbs, image.Path)
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

	uploadRoot := s.staticConfig.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 上传目录解析失败")
	}
	// 批量删除前先校验上传根目录节点本身，避免根目录被替换为符号链接。
	if err := pathpkg.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		log.Printf("BatchDeleteImages upload root security check failed: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}

	for _, img := range images {
		userSizeMap[img.UserID] += img.Size
		imageIDs = append(imageIDs, img.ID)
		fullPath, secureErr := pathpkg.SecureJoin(uploadRootAbs, img.Path)
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

// ListImages 分页查询图片列表；查询范围与过滤条件由上层传参决定。
func (s *ImageService) ListImages(params moduledto.ListImagesRequest) ([]model.Image, int64, int, int, error) {
	page, pageSize := normalizePagination(params.Page, params.PageSize)
	offset := (page - 1) * pageSize

	images, total, err := s.imageStore.ListImages(repo.ListImagesParams{
		UserID:      params.UserID,
		Username:    params.Username,
		Filename:    params.Filename,
		ID:          params.ID,
		Offset:      offset,
		Limit:       pageSize,
		PreloadUser: params.PreloadUser,
	})
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

// GetImageByID 获取指定图片；当 userID 非空时只在该用户范围内查询。
func (s *ImageService) GetImageByID(imageID uint, userID *uint) (*model.Image, error) {
	var (
		image *model.Image
		err   error
	)
	if userID != nil {
		image, err = s.imageStore.FindByIDAndUserID(imageID, *userID)
	} else {
		image, err = s.imageStore.FindByID(imageID)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, commonpkg.NewNotFoundError("图片不存在或无权访问")
		}
		return nil, commonpkg.NewInternalError("查找图片失败")
	}
	return image, nil
}

// GetImagesByIDs 按 ID 列表获取图片；当 userID 非空时只在该用户范围内查询。
func (s *ImageService) GetImagesByIDs(ids []uint, userID *uint) ([]model.Image, error) {
	var (
		images []model.Image
		err    error
	)
	if userID != nil {
		images, err = s.imageStore.FindByIDsAndUserID(ids, *userID)
	} else {
		images, err = s.imageStore.FindByIDs(ids)
	}
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

	// 1. 删除头像目录
	// 头像存储结构: data/avatars/{userID}/filename
	avatarRoot := s.staticConfig.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve avatar root: %w", err)
	}
	// 先校验头像根目录节点本身，避免根目录直接是符号链接。
	if err := pathpkg.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		return fmt.Errorf("avatar root symlink risk: %w", err)
	}

	userAvatarDir, err := pathpkg.SecureJoin(avatarRootAbs, fmt.Sprintf("%d", userID))
	if err != nil {
		return fmt.Errorf("failed to build avatar dir: %w", err)
	}
	// 在执行 RemoveAll 前再做一次链路检查，确保目标目录链路未被并发替换为符号链接。
	if err := pathpkg.EnsureNoSymlinkBetween(avatarRootAbs, userAvatarDir); err != nil {
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

	uploadRoot := s.staticConfig.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve upload root: %w", err)
	}
	// 先校验上传根目录节点本身，避免根目录直接是符号链接。
	if err := pathpkg.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		return fmt.Errorf("upload root symlink risk: %w", err)
	}

	for _, img := range images {
		// 转换路径分隔符以适配当前系统 (DB中存储的是 web 格式 '/')
		localPath := filepath.FromSlash(img.Path)
		fullPath, secureErr := pathpkg.SecureJoin(uploadRootAbs, localPath)
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

// ProcessImageUpload 处理图片上传核心业务：校验、配额检查、入库。
//
//nolint:gocyclo
func (s *ImageService) ProcessImageUpload(file *multipart.FileHeader, uid uint, usedSize int64, quota int64) (*model.Image, string, error) {
	valid, ext, err := s.ValidateImageFile(file)
	if !valid {
		return nil, "", err
	}

	if usedSize+file.Size > quota {
		return nil, "", commonpkg.NewForbiddenError(fmt.Sprintf("存储空间不足，上传失败。当前已用: %d B, 剩余: %d B", usedSize, quota-usedSize))
	}

	now := time.Now()
	datePath := filepath.Join(now.Format("2006"), now.Format("01"), now.Format("02"))

	cfg := s.staticConfig
	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 上传目录解析失败")
	}
	if err := pathpkg.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		log.Printf("Upload root security check failed: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}
	fullDir, err := pathpkg.SecureJoin(uploadRootAbs, datePath)
	if err != nil {
		log.Printf("SecureJoin dir error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 非法存储目录")
	}

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 无法创建存储目录")
	}
	if err := pathpkg.EnsureNoSymlinkBetween(uploadRootAbs, fullDir); err != nil {
		log.Printf("Upload dir security check failed: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 存储目录存在符号链接风险")
	}

	newFilename := uuid.New().String() + ext
	dst, err := pathpkg.SecureJoin(fullDir, newFilename)
	if err != nil {
		log.Printf("SecureJoin dst error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 非法文件路径")
	}

	src, err := file.Open()
	if err != nil {
		return nil, "", commonpkg.NewInternalError("无法读取上传文件")
	}
	defer func() { _ = src.Close() }()

	imgCfg, _, err := image.DecodeConfig(src)
	if err != nil {
		return nil, "", commonpkg.NewValidationError("无法解析图片尺寸，请上传有效图片")
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 无法重置文件读取位置")
	}

	out, err := os.Create(dst)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 无法创建文件")
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, src); err != nil {
		return nil, "", commonpkg.NewInternalError("文件保存失败")
	}

	relativePath := filepath.ToSlash(filepath.Join(
		now.Format("2006"), now.Format("01"), now.Format("02"), newFilename))

	imageRecord := model.Image{
		Filename:   newFilename,
		Path:       relativePath,
		Size:       file.Size,
		Width:      imgCfg.Width,
		Height:     imgCfg.Height,
		UserID:     uid,
		UploadedAt: now.Unix(),
		MimeType:   ext,
	}

	if err := s.imageStore.CreateAndIncreaseUserStorage(&imageRecord, uid, file.Size); err != nil {
		_ = os.Remove(dst)
		log.Printf("Process upload DB error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 数据库记录失败")
	}

	return &imageRecord, cfg.Upload.URLPrefix + relativePath, nil
}
