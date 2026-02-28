package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"perfect-pic-server/internal/common/httpx"
	"perfect-pic-server/internal/config"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/testutils"
	"perfect-pic-server/internal/utils"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	testService     *Service
	testUserService *Service
)

// Service 是测试专用聚合器，兼容旧测试中对单一 service 对象的调用方式。
type Service struct {
	dbConfig       *config.DBConfig
	authService    *AuthService
	userService    *UserService
	imageService   *ImageService
	emailService   *EmailService
	captchaService *CaptchaService
	initService    *InitService
	passkeyService *PasskeyService
	userStore      repository.UserStore
	passkeyStore   repository.PasskeyStore
}

type AuthError = httpx.AuthError
type AuthErrorCode = httpx.AuthErrorCode

const (
	AuthErrorValidation   = httpx.AuthErrorValidation
	AuthErrorUnauthorized = httpx.AuthErrorUnauthorized
	AuthErrorForbidden    = httpx.AuthErrorForbidden
	AuthErrorConflict     = httpx.AuthErrorConflict
	AuthErrorNotFound     = httpx.AuthErrorNotFound
	AuthErrorInternal     = httpx.AuthErrorInternal
)

func AsAuthError(err error) (*AuthError, bool) {
	return httpx.AsAuthError(err)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)
	passkeyStore := repository.NewPasskeyRepository(gdb)
	dbConfig := config.NewDBConfig(settingStore)

	authService := NewAuthService(dbConfig)
	userService := NewUserService(userStore, dbConfig)
	imageService := NewImageService(imageStore, dbConfig)
	emailService := NewEmailService(dbConfig)
	captchaService := NewCaptchaService(dbConfig)
	initService := NewInitService(systemStore, dbConfig)
	passkeyService := NewPasskeyService(passkeyStore, dbConfig)

	testService = &Service{
		dbConfig:       dbConfig,
		authService:    authService,
		userService:    userService,
		imageService:   imageService,
		emailService:   emailService,
		captchaService: captchaService,
		initService:    initService,
		passkeyService: passkeyService,
		userStore:      userStore,
		passkeyStore:   passkeyStore,
	}
	testUserService = testService

	if err := testService.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	testService.ClearCache()
	return gdb
}

func mustTestService(t *testing.T) *Service {
	t.Helper()
	if testService == nil {
		setupTestDB(t)
	}
	return testService
}

func (s *Service) InitializeSettings() error {
	return s.dbConfig.InitializeSettings()
}

func (s *Service) ClearCache() {
	s.dbConfig.ClearCache()
}

func (s *Service) IsSystemInitialized() bool {
	return s.initService.IsSystemInitialized()
}

func (s *Service) InitializeSystem(payload moduledto.InitRequest) error {
	return s.initService.InitializeSystem(payload)
}

func (s *Service) LoginUser(username, password string) (string, error) {
	user, err := s.userStore.FindByUsername(username)
	if err != nil {
		return "", httpx.NewAuthError(httpx.AuthErrorUnauthorized, "用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", httpx.NewAuthError(httpx.AuthErrorUnauthorized, "用户名或密码错误")
	}
	return s.authService.IssueLoginToken(user)
}

func (s *Service) RegisterUser(username, password, email string) error {
	if !s.initService.IsSystemInitialized() {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "系统尚未初始化，请先完成初始化")
	}
	if ok, msg := utils.ValidatePassword(password); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}
	if ok, msg := utils.ValidateUsername(username); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}
	if ok, msg := utils.ValidateEmail(email); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}
	if !s.dbConfig.GetBool(consts.ConfigAllowRegister) {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "注册功能已关闭")
	}
	usernameTaken, err := s.userService.IsUsernameTaken(username, nil, true)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}
	if usernameTaken {
		return httpx.NewAuthError(httpx.AuthErrorConflict, "用户名已存在")
	}
	emailTaken, err := s.userService.IsEmailTaken(email, nil, true)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}
	if emailTaken {
		return httpx.NewAuthError(httpx.AuthErrorConflict, "邮箱已被注册")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码加密失败")
	}
	newUser := model.User{
		Username:      username,
		Password:      string(hashedPassword),
		Email:         email,
		EmailVerified: false,
		Admin:         false,
		Avatar:        "",
	}
	if err := s.userService.CreateUser(&newUser); err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}
	verifyToken, err := utils.GenerateEmailToken(newUser.ID, newUser.Email, 30*time.Minute)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}
	baseURL := s.dbConfig.GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	verifyURL := fmt.Sprintf("%s/auth/email-verify?token=%s", baseURL, verifyToken)
	if s.emailService.ShouldSendEmail() {
		go func() {
			_ = s.emailService.SendVerificationEmail(newUser.Email, newUser.Username, verifyURL)
		}()
	}
	return nil
}

func (s *Service) VerifyEmail(token string) (bool, error) {
	claims, err := utils.ParseEmailToken(token)
	if err != nil {
		return false, httpx.NewAuthError(httpx.AuthErrorValidation, "验证链接已失效或不正确")
	}
	if claims.Type != "email_verify" {
		return false, httpx.NewAuthError(httpx.AuthErrorValidation, "无效的验证 Token 类型")
	}
	user, err := s.userStore.FindByID(claims.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, httpx.NewAuthError(httpx.AuthErrorNotFound, "用户不存在")
		}
		return false, httpx.NewAuthError(httpx.AuthErrorInternal, "验证失败，请稍后重试")
	}
	if user.Email != claims.Email {
		return false, httpx.NewAuthError(httpx.AuthErrorValidation, "邮箱不匹配，请重新发起验证")
	}
	if user.EmailVerified {
		return true, nil
	}
	user.EmailVerified = true
	if err := s.userService.SaveUser(user); err != nil {
		return false, httpx.NewAuthError(httpx.AuthErrorInternal, "验证失败，请稍后重试")
	}
	return false, nil
}

func (s *Service) VerifyEmailChange(token string) error {
	payload, ok := s.userService.VerifyEmailChangeToken(token)
	if !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, "验证链接已失效或不正确")
	}
	user, err := s.userStore.FindByID(payload.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.NewAuthError(httpx.AuthErrorNotFound, "用户不存在")
		}
		return httpx.NewAuthError(httpx.AuthErrorInternal, "邮箱修改失败，请稍后重试")
	}
	if user.Email != payload.OldEmail {
		return httpx.NewAuthError(httpx.AuthErrorValidation, "您的当前邮箱已变更，该验证链接已失效")
	}
	excludeID := payload.UserID
	emailTaken, err := s.userService.IsEmailTaken(payload.NewEmail, &excludeID, true)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "邮箱修改失败，请稍后重试")
	}
	if emailTaken {
		return httpx.NewAuthError(httpx.AuthErrorConflict, "新邮箱已被其他用户占用，无法修改")
	}
	user.Email = payload.NewEmail
	user.EmailVerified = true
	if err := s.userService.SaveUser(user); err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "邮箱修改失败，请稍后重试")
	}
	return nil
}

func (s *Service) RequestPasswordReset(email string) error {
	user, err := s.userStore.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return httpx.NewAuthError(httpx.AuthErrorInternal, "生成重置链接失败，请稍后重试")
	}
	if user.Status == 2 || user.Status == 3 {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "该账号已被封禁或停用，无法重置密码")
	}
	token, err := s.userService.GenerateForgetPasswordToken(user.ID)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "生成重置链接失败，请稍后重试")
	}
	baseURL := s.dbConfig.GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", baseURL, token)
	if s.emailService.ShouldSendEmail() {
		go func() {
			_ = s.emailService.SendPasswordResetEmail(user.Email, user.Username, resetURL)
		}()
	}
	return nil
}

func (s *Service) ResetPassword(token, newPassword string) error {
	if ok, msg := utils.ValidatePassword(newPassword); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}
	userID, valid := s.userService.VerifyForgetPasswordToken(token)
	if !valid {
		return httpx.NewAuthError(httpx.AuthErrorValidation, "重置链接无效或已过期")
	}
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.NewAuthError(httpx.AuthErrorNotFound, "用户不存在")
		}
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码重置失败")
	}
	if user.Status == 2 || user.Status == 3 {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "该账号已被封禁或停用")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码加密失败")
	}
	user.Password = string(hashedPassword)
	user.EmailVerified = true
	if err := s.userService.SaveUser(user); err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码重置失败")
	}
	return nil
}

func (s *Service) SendTestEmail(toEmail string) error {
	return s.emailService.SendTestEmail(toEmail)
}

func (s *Service) SendVerificationEmail(toEmail, username, verifyURL string) error {
	return s.emailService.SendVerificationEmail(toEmail, username, verifyURL)
}

func (s *Service) SendEmailChangeVerification(toEmail, username, oldEmail, newEmail, verifyURL string) error {
	return s.emailService.SendEmailChangeVerification(toEmail, username, oldEmail, newEmail, verifyURL)
}

func (s *Service) SendPasswordResetEmail(toEmail, username, resetURL string) error {
	return s.emailService.SendPasswordResetEmail(toEmail, username, resetURL)
}

func (s *Service) GetCaptchaProviderInfo() moduledto.CaptchaProviderResponse {
	return s.captchaService.GetCaptchaProviderInfo()
}

func (s *Service) VerifyCaptchaChallenge(captchaID, captchaAnswer, captchaToken, remoteIP string) (bool, string) {
	return s.captchaService.VerifyCaptchaChallenge(captchaID, captchaAnswer, captchaToken, remoteIP)
}

func (s *Service) ValidateImageFile(file *multipart.FileHeader) (bool, string, error) {
	return s.imageService.ValidateImageFile(file)
}

func (s *Service) ProcessImageUpload(file *multipart.FileHeader, uid uint) (*model.Image, string, error) {
	valid, ext, err := s.imageService.ValidateImageFile(file)
	if !valid {
		return nil, "", err
	}
	user, err := s.userStore.FindByID(uid)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("查询用户信息失败")
	}
	usedSize := user.StorageUsed
	quota := s.dbConfig.GetDefaultStorageQuota()
	if user.StorageQuota != nil {
		quota = *user.StorageQuota
	}
	if usedSize+file.Size > quota {
		return nil, "", commonpkg.NewForbiddenError(fmt.Sprintf("存储空间不足，上传失败。当前已用: %d B, 剩余: %d B", usedSize, quota-usedSize))
	}
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
	if err := utils.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 上传目录存在符号链接风险")
	}
	fullDir, err := utils.SecureJoin(uploadRootAbs, datePath)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 非法存储目录")
	}
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 无法创建存储目录")
	}
	if err := utils.EnsureNoSymlinkBetween(uploadRootAbs, fullDir); err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 存储目录存在符号链接风险")
	}
	newFilename := uuid.New().String() + ext
	dst, err := utils.SecureJoin(fullDir, newFilename)
	if err != nil {
		return nil, "", commonpkg.NewInternalError("系统错误: 非法文件路径")
	}
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
	if err := s.imageService.CreateAndIncreaseUserStorage(&imageRecord, uid, file.Size); err != nil {
		_ = os.Remove(dst)
		return nil, "", commonpkg.NewInternalError("系统错误: 数据库记录失败")
	}
	return &imageRecord, cfg.Upload.URLPrefix + relativePath, nil
}

func (s *Service) DeleteImage(image *model.Image) error {
	return s.imageService.DeleteImage(image)
}

func (s *Service) BatchDeleteImages(images []model.Image) error {
	return s.imageService.BatchDeleteImages(images)
}

func (s *Service) UpdateUserAvatar(user *model.User, file *multipart.FileHeader) (string, error) {
	cfg := config.Get()
	avatarRoot := cfg.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return "", commonpkg.NewInternalError("系统错误: 头像目录解析失败")
	}
	if err := utils.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		return "", commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	userIDStr := fmt.Sprintf("%v", user.ID)
	storageDir, err := utils.SecureJoin(avatarRootAbs, userIDStr)
	if err != nil {
		return "", commonpkg.NewInternalError("系统错误: 非法头像目录")
	}
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", commonpkg.NewInternalError("系统错误: 无法创建存储目录")
	}
	if err := utils.EnsureNoSymlinkBetween(avatarRootAbs, storageDir); err != nil {
		return "", commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}
	newFilename := uuid.New().String() + ext
	dstPath, err := utils.SecureJoin(storageDir, newFilename)
	if err != nil {
		return "", commonpkg.NewInternalError("系统错误: 非法头像文件路径")
	}
	src, err := file.Open()
	if err != nil {
		return "", commonpkg.NewInternalError("无法读取上传文件")
	}
	defer func() { _ = src.Close() }()
	out, err := os.Create(dstPath)
	if err != nil {
		return "", commonpkg.NewInternalError("系统错误: 无法创建文件")
	}
	defer func() { _ = out.Close() }()
	if _, err = io.Copy(out, src); err != nil {
		return "", commonpkg.NewInternalError("文件保存失败")
	}
	oldAvatar := user.Avatar
	if err := s.userService.UpdateAvatar(user, newFilename); err != nil {
		_ = os.Remove(dstPath)
		return "", commonpkg.NewInternalError("系统错误: 数据库更新失败")
	}
	if oldAvatar != "" {
		oldAvatarPath, secureErr := utils.SecureJoin(storageDir, oldAvatar)
		if secureErr == nil {
			_ = os.Remove(oldAvatarPath)
		}
	}
	return newFilename, nil
}

func (s *Service) RemoveUserAvatar(user *model.User) error {
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
	if err := utils.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		return commonpkg.NewInternalError("系统错误: 头像目录存在符号链接风险")
	}
	userIDStr := fmt.Sprintf("%v", user.ID)
	storageDir, err := utils.SecureJoin(avatarRootAbs, userIDStr)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 非法头像目录")
	}
	oldAvatarPath, err := utils.SecureJoin(storageDir, user.Avatar)
	if err != nil {
		return commonpkg.NewInternalError("系统错误: 非法头像文件路径")
	}
	if err := s.userService.ClearAvatar(user); err != nil {
		return commonpkg.NewInternalError("系统错误: 移除头像失败")
	}
	if err := os.Remove(oldAvatarPath); err != nil && !os.IsNotExist(err) {
		return commonpkg.NewInternalError("系统错误: 移除头像失败")
	}
	return nil
}

func (s *Service) ListUserImages(params moduledto.UserImageListRequest) ([]model.Image, int64, int, int, error) {
	return s.imageService.ListUserImages(params)
}

func (s *Service) GetUserImageCount(userID uint) (int64, error) {
	return s.imageService.GetUserImageCount(userID)
}

func (s *Service) GetUserOwnedImage(imageID uint, userID uint) (*model.Image, error) {
	return s.imageService.GetUserOwnedImage(imageID, userID)
}

func (s *Service) GetImagesByIDsForUser(ids []uint, userID uint) ([]model.Image, error) {
	return s.imageService.GetImagesByIDsForUser(ids, userID)
}

func (s *Service) DeleteUserFiles(userID uint) error {
	return s.imageService.DeleteUserFiles(userID)
}

func (s *Service) AdminListImages(params moduledto.AdminImageListRequest) ([]model.Image, int64, int, int, error) {
	return s.imageService.AdminListImages(params)
}

func (s *Service) AdminGetImageByID(id uint) (*model.Image, error) {
	return s.imageService.AdminGetImageByID(id)
}

func (s *Service) AdminGetImagesByIDs(ids []uint) ([]model.Image, error) {
	return s.imageService.AdminGetImagesByIDs(ids)
}

func (s *Service) GenerateForgetPasswordToken(userID uint) (string, error) {
	return s.userService.GenerateForgetPasswordToken(userID)
}

func (s *Service) VerifyForgetPasswordToken(token string) (uint, bool) {
	return s.userService.VerifyForgetPasswordToken(token)
}

func (s *Service) GenerateEmailChangeToken(userID uint, oldEmail, newEmail string) (string, error) {
	return s.userService.GenerateEmailChangeToken(userID, oldEmail, newEmail)
}

func (s *Service) VerifyEmailChangeToken(token string) (*moduledto.EmailChangeToken, bool) {
	return s.userService.VerifyEmailChangeToken(token)
}

func (s *Service) GetSystemDefaultStorageQuota() int64 {
	return s.userService.GetSystemDefaultStorageQuota()
}

func (s *Service) AdminGetUserDetail(id uint) (*model.User, error) {
	return s.userService.AdminGetUserDetail(id)
}

func (s *Service) AdminListUsers(params moduledto.AdminUserListRequest) ([]model.User, int64, error) {
	return s.userService.AdminListUsers(params)
}

func (s *Service) AdminCreateUser(input moduledto.AdminCreateUserRequest) (*model.User, error) {
	return s.userService.AdminCreateUser(input)
}

func (s *Service) AdminPrepareUserUpdates(userID uint, req moduledto.AdminUserUpdateRequest) (map[string]interface{}, error) {
	return s.userService.AdminPrepareUserUpdates(userID, req)
}

func (s *Service) AdminApplyUserUpdates(userID uint, updates map[string]interface{}) error {
	return s.userService.AdminApplyUserUpdates(userID, updates)
}

func (s *Service) AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		if err := s.imageService.DeleteUserFiles(userID); err != nil {
			return commonpkg.NewInternalError("删除用户失败")
		}
		if err := s.passkeyService.DeletePasskeyCredentialsByUserID(userID); err != nil {
			return commonpkg.NewInternalError("删除用户失败")
		}
		if err := s.userService.HardDeleteUserWithImages(userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return commonpkg.NewNotFoundError("用户不存在")
			}
			return commonpkg.NewInternalError("删除用户失败")
		}
		return nil
	}
	if err := s.userService.SoftDeleteUser(userID, time.Now().Unix()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return commonpkg.NewNotFoundError("用户不存在")
		}
		return commonpkg.NewInternalError("删除用户失败")
	}
	return nil
}

func (s *Service) IsUsernameTaken(username string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	return s.userService.IsUsernameTaken(username, excludeUserID, includeDeleted)
}

func (s *Service) IsEmailTaken(email string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	return s.userService.IsEmailTaken(email, excludeUserID, includeDeleted)
}

func (s *Service) GetUserProfile(userID uint) (*moduledto.UserProfileResponse, error) {
	return s.userService.GetUserProfile(userID)
}

func (s *Service) UpdateUsernameAndGenerateToken(userID uint, newUsername string, isAdmin bool) (string, error) {
	return s.userService.UpdateUsernameAndGenerateToken(userID, newUsername, isAdmin)
}

func (s *Service) UpdatePasswordByOldPassword(userID uint, oldPassword, newPassword string) error {
	return s.userService.UpdatePasswordByOldPassword(userID, oldPassword, newPassword)
}

func (s *Service) RequestEmailChange(userID uint, password, newEmail string) error {
	if ok, msg := utils.ValidateEmail(newEmail); !ok {
		return commonpkg.NewValidationError(msg)
	}
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		return commonpkg.NewNotFoundError("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return commonpkg.NewForbiddenError("密码错误")
	}
	if user.Email == newEmail {
		return commonpkg.NewValidationError("新邮箱不能与当前邮箱相同")
	}
	emailTaken, err := s.userService.IsEmailTaken(newEmail, nil, true)
	if err != nil {
		return commonpkg.NewInternalError("生成验证链接失败")
	}
	if emailTaken {
		return commonpkg.NewConflictError("该邮箱已被使用")
	}
	token, err := s.userService.GenerateEmailChangeToken(user.ID, user.Email, newEmail)
	if err != nil {
		return commonpkg.NewInternalError("生成验证链接失败")
	}
	baseURL := s.dbConfig.GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	verifyURL := fmt.Sprintf("%s/auth/email-change-verify?token=%s", baseURL, token)
	if s.emailService.ShouldSendEmail() {
		go func() {
			_ = s.emailService.SendEmailChangeVerification(newEmail, user.Username, user.Email, newEmail, verifyURL)
		}()
	}
	return nil
}

func (s *Service) BeginPasskeyRegistration(userID uint) (string, *protocol.CredentialCreation, error) {
	if err := s.ensureUserPasskeyCapacity(userID); err != nil {
		return "", nil, err
	}
	webauthnClient, err := s.passkeyService.CreatePasskeyWebAuthnClient()
	if err != nil {
		return "", nil, err
	}
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, commonpkg.NewNotFoundError("用户不存在")
		}
		return "", nil, commonpkg.NewInternalError("读取用户信息失败")
	}
	credentials, err := s.passkeyService.LoadUserPasskeyCredentials(userID)
	if err != nil {
		return "", nil, err
	}
	passkeyUser := &testPasskeyWebAuthnUser{
		userID:      userID,
		username:    user.Username,
		id:          []byte(strconv.FormatUint(uint64(userID), 10)),
		credentials: credentials,
	}
	creation, sessionData, err := webauthnClient.BeginRegistration(
		passkeyUser,
		webauthn.WithCredentialParameters(s.passkeyService.GetPasskeyRecommendedCredentialParameters()),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(passkeyUser.credentials).CredentialDescriptors()),
		webauthn.WithExtensions(protocol.AuthenticationExtensions{"credProps": true}),
	)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 注册挑战失败")
	}
	sessionID, err := s.passkeyService.StorePasskeySession(consts.PasskeySessionRegistration, userID, sessionData)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 注册会话失败")
	}
	return sessionID, creation, nil
}

func (s *Service) FinishPasskeyRegistration(userID uint, sessionID string, credentialJSON []byte) error {
	_, err := s.passkeyService.ConsumePasskeyRegistrationSession(sessionID, userID)
	if err != nil {
		return err
	}
	_ = credentialJSON
	return commonpkg.NewValidationError("Passkey 注册校验失败，请重试")
}

func (s *Service) BeginPasskeyLogin() (string, *protocol.CredentialAssertion, error) {
	webauthnClient, err := s.passkeyService.CreatePasskeyWebAuthnClient()
	if err != nil {
		return "", nil, err
	}
	assertion, sessionData, err := webauthnClient.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 登录挑战失败")
	}
	sessionID, err := s.passkeyService.StorePasskeySession(consts.PasskeySessionLogin, 0, sessionData)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 登录会话失败")
	}
	return sessionID, assertion, nil
}

func (s *Service) FinishPasskeyLogin(sessionID string, credentialJSON []byte) (string, error) {
	_, err := s.passkeyService.ConsumePasskeyLoginSession(sessionID)
	if err != nil {
		return "", err
	}
	_ = credentialJSON
	return "", commonpkg.NewUnauthorizedError("Passkey 登录失败")
}

func (s *Service) ListUserPasskeys(userID uint) ([]moduledto.UserPasskeyResponse, error) {
	return s.passkeyService.ListUserPasskeys(userID)
}

func (s *Service) DeleteUserPasskey(userID uint, passkeyID uint) error {
	return s.passkeyService.DeleteUserPasskey(userID, passkeyID)
}

func (s *Service) UpdateUserPasskeyName(userID uint, passkeyID uint, name string) error {
	return s.passkeyService.UpdateUserPasskeyName(userID, passkeyID, name)
}

type testPasskeyWebAuthnUser struct {
	userID      uint
	username    string
	id          []byte
	credentials []webauthn.Credential
}

func (u *testPasskeyWebAuthnUser) WebAuthnID() []byte {
	return u.id
}

func (u *testPasskeyWebAuthnUser) WebAuthnName() string {
	return u.username
}

func (u *testPasskeyWebAuthnUser) WebAuthnDisplayName() string {
	return u.username
}

func (u *testPasskeyWebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (s *Service) ensureUserPasskeyCapacity(userID uint) error {
	count, err := s.passkeyStore.CountPasskeyCredentialsByUserID(userID)
	if err != nil {
		return commonpkg.NewInternalError("校验 Passkey 数量失败")
	}
	if count >= consts.MaxUserPasskeyCount {
		return commonpkg.NewConflictError("Passkey 数量已达上限（最多 10 个）")
	}
	return nil
}

func resetPasswordResetStore() {
	clearSyncMap(&passwordResetStore)
	clearSyncMap(&passwordResetTokenStore)
}

func resetEmailChangeStore() {
	clearSyncMap(&emailChangeStore)
	clearSyncMap(&emailChangeTokenStore)
}

func resetPasskeySessionStore() {
	clearSyncMap(&passkeySessionStore)
}

func clearSyncMap(store *sync.Map) {
	store.Range(func(key, _ interface{}) bool {
		store.Delete(key)
		return true
	})
}
