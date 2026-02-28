package app

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProcessImageUpload 处理图片上传核心业务
// 包括：配额检查、文件保存、数据库记录
//
//nolint:gocyclo
func (c *ImageUseCase) ProcessImageUpload(file *multipart.FileHeader, uid uint) (*model.Image, string, error) {
	// 验证文件
	valid, ext, err := c.imageService.ValidateImageFile(file)
	if !valid {
		return nil, "", err
	}

	// 检查配额 (使用 StorageUsed 字段)
	user, err := c.userStore.FindByID(uid)
	if err != nil {
		log.Printf("Get user error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("查询用户信息失败")
	}

	// 如果 StorageUsed 为 0 但不是新用户，可能需要同步
	usedSize := user.StorageUsed

	var quota int64
	if user.StorageQuota != nil {
		quota = *user.StorageQuota
	} else {
		quota = c.dbConfig.GetDefaultStorageQuota()
	}

	if usedSize+file.Size > quota {
		return nil, "", commonpkg.NewForbiddenError(fmt.Sprintf("存储空间不足，上传失败。当前已用: %d B, 剩余: %d B", usedSize, quota-usedSize))
	}

	// 准备路径
	now := time.Now()
	datePath := filepath.Join(now.Format("2006"), now.Format("01"), now.Format("02"))

	cfg := config.Get()
	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 上传目录解析失败")
	}
	// 先检查上传根目录节点本身不是符号链接（防止根目录直接指向外部路径）。
	if err := utils.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		log.Printf("Upload root security check failed: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}
	// 完整的磁盘文件夹路径
	fullDir, err := utils.SecureJoin(uploadRootAbs, datePath)
	if err != nil {
		log.Printf("SecureJoin dir error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 非法存储目录")
	}

	// 自动创建文件夹
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 无法创建存储目录")
	}
	// 目录创建后再次检查链路，降低 TOCTOU 风险。
	if err := utils.EnsureNoSymlinkBetween(uploadRootAbs, fullDir); err != nil {
		log.Printf("Upload dir security check failed: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 存储目录存在符号链接风险")
	}

	// 生成唯一文件名
	newFilename := uuid.New().String() + ext
	dst, err := utils.SecureJoin(fullDir, newFilename)
	if err != nil {
		log.Printf("SecureJoin dst error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 非法文件路径")
	}

	// 保存文件 (IO 操作放在事务前，如果 DB 失败则删除文件)
	src, err := file.Open()
	if err != nil {
		return nil, "", commonpkg.NewInternalError("无法读取上传文件")
	}
	defer func() { _ = src.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 无法创建文件")
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, src); err != nil {
		return nil, "", commonpkg.NewInternalError("文件保存失败")
	}

	// 数据库操作 (事务)
	relativePath := filepath.ToSlash(filepath.Join(
		now.Format("2006"), now.Format("01"), now.Format("02"), newFilename))

	imageRecord := model.Image{
		Filename:   newFilename,
		Path:       relativePath,
		Size:       file.Size,
		UserID:     uid,
		UploadedAt: now.Unix(),
		MimeType:   ext,
	}

	if err := c.imageService.CreateAndIncreaseUserStorage(&imageRecord, uid, file.Size); err != nil {
		_ = os.Remove(dst) // 回滚文件
		log.Printf("Process upload DB error: %v\n", err)
		return nil, "", commonpkg.NewInternalError("系统错误: 数据库记录失败")
	}

	return &imageRecord, cfg.Upload.URLPrefix + relativePath, nil
}

// UpdateUserAvatar 更新用户头像
func (c *ImageUseCase) UpdateUserAvatar(user *model.User, file *multipart.FileHeader) (string, error) {
	cfg := config.Get()
	avatarRoot := cfg.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return "", commonpkg.NewInternalError("系统错误: 头像目录解析失败")
	}
	// 先检查头像根目录节点本身不是符号链接。
	if err := utils.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		log.Printf("Avatar root security check failed: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	// userIdStr for path
	userIdStr := fmt.Sprintf("%v", user.ID)
	storageDir, err := utils.SecureJoin(avatarRootAbs, userIdStr)
	if err != nil {
		log.Printf("Avatar storage dir error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 非法头像目录")
	}

	// 自动创建文件夹
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 无法创建存储目录")
	}
	// 用户目录创建后再次检查链路，确保新增层级未被符号链接替换。
	if err := utils.EnsureNoSymlinkBetween(avatarRootAbs, storageDir); err != nil {
		log.Printf("Avatar storage security check failed: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}

	// 生成唯一文件名
	newFilename := uuid.New().String() + ext
	dstPath, err := utils.SecureJoin(storageDir, newFilename)
	if err != nil {
		log.Printf("Avatar dst secure join error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 非法头像文件路径")
	}

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		log.Printf("File open error: %v\n", err)
		return "", commonpkg.NewInternalError("无法读取上传文件")
	}
	defer func() { _ = src.Close() }()

	// 创建目标文件
	out, err := os.Create(dstPath)
	if err != nil {
		log.Printf("File create error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 无法创建文件")
	}
	defer func() { _ = out.Close() }()

	// 复制内容
	if _, err = io.Copy(out, src); err != nil {
		log.Printf("File save error: %v\n", err)
		return "", commonpkg.NewInternalError("文件保存失败")
	}

	// 保存旧头像文件名用于后续删除
	oldAvatar := user.Avatar

	// 更新数据库
	if err := c.userService.UpdateAvatar(user, newFilename); err != nil {
		_ = os.Remove(dstPath) // 回滚文件
		log.Printf("DB Update avatar error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 数据库更新失败")
	}

	// 删除旧头像
	if oldAvatar != "" {
		oldAvatarPath, secureErr := utils.SecureJoin(storageDir, oldAvatar)
		if secureErr != nil {
			log.Printf("Old avatar secure path error: %v\n", secureErr)
		} else {
			_ = os.Remove(oldAvatarPath)
		}
	}

	return newFilename, nil
}

// RemoveUserAvatar 移除用户头像
func (c *ImageUseCase) RemoveUserAvatar(user *model.User) error {
	// 如果用户没有头像，直接返回
	if user.Avatar == "" {
		return nil
	}

	cfg := config.Get()
	avatarRoot := cfg.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 头像目录解析失败")
	}
	// 删除头像前先校验头像根目录节点本身，防止根目录符号链接穿透。
	if err := utils.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		log.Printf("RemoveUserAvatar root security check failed: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}

	userIdStr := fmt.Sprintf("%v", user.ID)
	storageDir, err := utils.SecureJoin(avatarRootAbs, userIdStr)
	if err != nil {
		log.Printf("RemoveUserAvatar storage dir secure join error: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 非法头像目录")
	}
	oldAvatarPath, err := utils.SecureJoin(storageDir, user.Avatar)
	if err != nil {
		log.Printf("RemoveUserAvatar file secure join error: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 非法头像文件路径")
	}

	// 更新数据库
	if err := c.userService.ClearAvatar(user); err != nil {
		log.Printf("DB Remove avatar error: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 移除头像失败")
	}

	// 删除文件
	if err := os.Remove(oldAvatarPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Remove avatar file error: %v\n", err)
	}

	return nil
}