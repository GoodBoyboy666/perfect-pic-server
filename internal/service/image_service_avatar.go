package service

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/pkg/pathpkg"
	"strings"

	"github.com/google/uuid"
)

// SaveUserAvatarFile 保存用户头像文件并返回新文件名。
func (s *ImageService) SaveUserAvatarFile(userID uint, file *multipart.FileHeader) (string, error) {
	avatarRootAbs, storageDir, err := s.resolveUserAvatarDir(userID)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Printf("MkdirAll error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 无法创建存储目录")
	}
	if err := pathpkg.EnsureNoSymlinkBetween(avatarRootAbs, storageDir); err != nil {
		log.Printf("Avatar storage security check failed: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	newFilename := uuid.New().String() + ext
	dstPath, err := pathpkg.SecureJoin(storageDir, newFilename)
	if err != nil {
		log.Printf("Avatar dst secure join error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 非法头像文件路径")
	}

	src, err := file.Open()
	if err != nil {
		log.Printf("File open error: %v\n", err)
		return "", commonpkg.NewInternalError("无法读取上传文件")
	}
	defer func() { _ = src.Close() }()

	out, err := os.Create(dstPath)
	if err != nil {
		log.Printf("File create error: %v\n", err)
		return "", commonpkg.NewInternalError("系统错误: 无法创建文件")
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, src); err != nil {
		log.Printf("File save error: %v\n", err)
		return "", commonpkg.NewInternalError("文件保存失败")
	}

	return newFilename, nil
}

// DeleteUserAvatarFile 删除指定用户的头像文件（不存在时静默成功）。
func (s *ImageService) DeleteUserAvatarFile(userID uint, filename string) error {
	if filename == "" {
		return nil
	}

	_, storageDir, err := s.resolveUserAvatarDir(userID)
	if err != nil {
		return err
	}

	avatarPath, err := pathpkg.SecureJoin(storageDir, filename)
	if err != nil {
		log.Printf("Avatar file secure join error: %v\n", err)
		return commonpkg.NewInternalError("系统错误: 非法头像文件路径")
	}

	if err := os.Remove(avatarPath); err != nil && !os.IsNotExist(err) {
		return commonpkg.NewInternalError("系统错误: 删除头像文件失败")
	}
	return nil
}

func (s *ImageService) resolveUserAvatarDir(userID uint) (string, string, error) {
	avatarRoot := s.staticConfig.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return "", "", commonpkg.NewInternalError("系统错误: 头像目录解析失败")
	}
	if err := pathpkg.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		log.Printf("Avatar root security check failed: %v\n", err)
		return "", "", commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}

	storageDir, err := pathpkg.SecureJoin(avatarRootAbs, fmt.Sprintf("%v", userID))
	if err != nil {
		log.Printf("Avatar storage dir error: %v\n", err)
		return "", "", commonpkg.NewInternalError("系统错误: 非法头像目录")
	}

	return avatarRootAbs, storageDir, nil
}
