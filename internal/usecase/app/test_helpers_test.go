package app

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/common/httpx"
	"perfect-pic-server/internal/config"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/pkg/cache"
	pkgmail "perfect-pic-server/internal/pkg/email"
	jwtpkg "perfect-pic-server/internal/pkg/jwt"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"
	"testing"

	"gorm.io/gorm"
)

type appFixture struct {
	gdb            *gorm.DB
	dbConfig       *config.DBConfig
	userStore      repository.UserStore
	passkeyStore   repository.PasskeyStore
	authService    *service.AuthService
	userService    *service.UserService
	imageService   *service.ImageService
	emailService   *service.EmailService
	captchaService *service.CaptchaService
	initService    *service.InitService
	passkeyService *service.PasskeyService
	authUC         *AuthUseCase
	userUC         *UserUseCase
	imageUC        *ImageUseCase
	passkeyUC      *PasskeyUseCase
}

var testGormDB *gorm.DB

func setupAppFixture(t *testing.T) *appFixture {
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
	if err := dbConfig.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	dbConfig.ClearCache()

	authService := service.NewAuthService(dbConfig, tokenService)
	userService := service.NewUserService(userStore, dbConfig, cacheStore, tokenService)
	imageService := service.NewImageService(imageStore, dbConfig, staticConfig)
	emailService := service.NewEmailService(dbConfig, pkgmail.NewMailer(), staticConfig)
	captchaService := service.NewCaptchaService(dbConfig)
	initService := service.NewInitService(systemStore, dbConfig)
	passkeyService := service.NewPasskeyService(passkeyStore, dbConfig, cacheStore)

	authUC := NewAuthUseCase(authService, userStore, userService, emailService, initService, dbConfig)
	userUC := NewUserUseCase(authService, userService, userStore, emailService, dbConfig)
	imageUC := NewImageUseCase(imageService, userService, userStore, staticConfig, dbConfig)
	passkeyUC := NewPasskeyUseCase(passkeyService, passkeyStore, authService, userStore)

	return &appFixture{
		gdb:            gdb,
		dbConfig:       dbConfig,
		userStore:      userStore,
		passkeyStore:   passkeyStore,
		authService:    authService,
		userService:    userService,
		imageService:   imageService,
		emailService:   emailService,
		captchaService: captchaService,
		initService:    initService,
		passkeyService: passkeyService,
		authUC:         authUC,
		userUC:         userUC,
		imageUC:        imageUC,
		passkeyUC:      passkeyUC,
	}
}

func (f *appFixture) initializeSystem(t *testing.T) {
	t.Helper()
	payload := moduledto.InitRequest{
		Username:        "admin_1",
		Password:        "abc12345",
		SiteName:        "TestSite",
		SiteDescription: "TestDesc",
	}
	if err := f.initService.InitializeSystem(payload); err != nil {
		t.Fatalf("InitializeSystem failed: %v", err)
	}
	f.dbConfig.ClearCache()
}

func assertAuthErrorCode(t *testing.T, err error, code httpx.AuthErrorCode) *httpx.AuthError {
	t.Helper()
	authErr, ok := httpx.AsAuthError(err)
	if !ok {
		t.Fatalf("expected AuthError, got: %v", err)
	}
	if authErr.Code != code {
		t.Fatalf("expected code=%q, got=%q", code, authErr.Code)
	}
	return authErr
}

func assertServiceErrorCode(t *testing.T, err error, code common.ErrorCode) *common.ServiceError {
	t.Helper()
	serviceErr, ok := common.AsServiceError(err)
	if !ok {
		t.Fatalf("expected ServiceError, got: %v", err)
	}
	if serviceErr.Code != code {
		t.Fatalf("expected code=%q, got=%q", code, serviceErr.Code)
	}
	return serviceErr
}

func mustFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(int64(len(content)) + 1024); err != nil {
		t.Fatalf("ParseMultipartForm failed: %v", err)
	}
	files := req.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected 1 file header, got %d", len(files))
	}
	return files[0]
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
}
