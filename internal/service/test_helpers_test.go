package service

import (
	"encoding/json"
	"mime/multipart"
	"strconv"
	"testing"
	"time"

	"perfect-pic-server/internal/config"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/pkg/cache"
	pkgmail "perfect-pic-server/internal/pkg/email"
	jwtpkg "perfect-pic-server/internal/pkg/jwt"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var (
	testService *Service
	testGormDB  *gorm.DB
)

// Service 是测试专用聚合器，仅做服务层直连转发，不复制业务编排逻辑。
type Service struct {
	dbConfig       *config.DBConfig
	authService    *AuthService
	userService    *UserService
	imageService   *ImageService
	emailService   *EmailService
	captchaService *CaptchaService
	initService    *InitService
	passkeyService *PasskeyService
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	testGormDB = gdb
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)
	passkeyStore := repository.NewPasskeyRepository(gdb)
	dbConfig := config.NewDBConfig(settingStore)
	staticConfig := config.NewStaticConfig()
	tokenService := jwtpkg.NewJWT(config.NewJWTConfig(staticConfig))
	cacheStore := cache.NewStore(nil, config.NewCacheConfig(staticConfig))

	authService := NewAuthService(dbConfig, tokenService)
	userService := NewUserService(userStore, dbConfig, cacheStore, tokenService)
	imageService := NewImageService(imageStore, dbConfig, staticConfig)
	emailService := NewEmailService(dbConfig, pkgmail.NewMailer(), staticConfig)
	captchaService := NewCaptchaService(dbConfig)
	initService := NewInitService(systemStore, dbConfig)
	passkeyService := NewPasskeyService(passkeyStore, dbConfig, cacheStore)

	testService = &Service{
		dbConfig:       dbConfig,
		authService:    authService,
		userService:    userService,
		imageService:   imageService,
		emailService:   emailService,
		captchaService: captchaService,
		initService:    initService,
		passkeyService: passkeyService,
	}

	if err := testService.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	testService.ClearCache()
	return gdb
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

func (s *Service) ValidateImageFile(fileHeader *multipart.FileHeader) (bool, string, error) {
	return s.imageService.ValidateImageFile(fileHeader)
}

func (s *Service) DeleteImage(image *model.Image) error {
	return s.imageService.DeleteImage(image)
}

func (s *Service) BatchDeleteImages(images []model.Image) error {
	return s.imageService.BatchDeleteImages(images)
}

func (s *Service) ListImages(params moduledto.ListImagesRequest) ([]model.Image, int64, int, int, error) {
	return s.imageService.ListImages(params)
}

func (s *Service) GetUserImageCount(userID uint) (int64, error) {
	return s.imageService.GetUserImageCount(userID)
}

func (s *Service) GetImageByID(imageID uint, userID *uint) (*model.Image, error) {
	return s.imageService.GetImageByID(imageID, userID)
}

func (s *Service) GetImagesByIDs(ids []uint, userID *uint) ([]model.Image, error) {
	return s.imageService.GetImagesByIDs(ids, userID)
}

func (s *Service) DeleteUserFiles(userID uint) error {
	return s.imageService.DeleteUserFiles(userID)
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

func (s *Service) GetUserByID(id uint, includeDeleted bool) (*model.User, error) {
	return s.userService.GetUserByID(id, includeDeleted)
}

func (s *Service) AdminGetUserDetail(id uint) (*model.User, error) {
	return s.userService.GetUserByID(id, true)
}

func (s *Service) AdminListUsers(params moduledto.UserListRequest) ([]model.User, int64, error) {
	return s.userService.ListUsers(params)
}

func (s *Service) CreateUser(input moduledto.CreateUserRequest, allowReservedUsername bool) (*model.User, error) {
	return s.userService.CreateUser(input, allowReservedUsername)
}

func (s *Service) UpdateUser(userID uint, req moduledto.UpdateUserRequest, allowReservedUsername bool) error {
	return s.userService.UpdateUser(userID, req, allowReservedUsername)
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

func (s *Service) UpdatePasswordByOldPassword(userID uint, oldPassword, newPassword string) error {
	return s.userService.UpdatePasswordByOldPassword(userID, oldPassword, newPassword)
}

func resetPasswordResetStore() {
	if testService == nil || testService.userService == nil {
		return
	}
	testService.userService.resetTokenCachesForTest()
}

func resetEmailChangeStore() {
	if testService == nil || testService.userService == nil {
		return
	}
	testService.userService.resetTokenCachesForTest()
}

func ttlForLocalToken(expiresAt time.Time) time.Duration {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return -time.Nanosecond
	}
	return ttl
}

func (s *UserService) resetTokenCachesForTest() {
	if s.cache != nil {
		s.cache.ClearLocal()
	}
}

func (s *UserService) putEmailChangeTokenLocalForTest(token moduledto.EmailChangeToken) {
	if s.cache == nil {
		return
	}

	payload, err := json.Marshal(moduledto.EmailChangeRedisPayload{
		UserID:   token.UserID,
		OldEmail: token.OldEmail,
		NewEmail: token.NewEmail,
	})
	if err != nil {
		return
	}

	s.cache.SetIndexed(
		s.cache.RedisKey("email_change", "user", strconv.FormatUint(uint64(token.UserID), 10)),
		s.cache.RedisKey("email_change", "token", token.Token),
		string(payload),
		ttlForLocalToken(token.ExpiresAt),
	)
}
