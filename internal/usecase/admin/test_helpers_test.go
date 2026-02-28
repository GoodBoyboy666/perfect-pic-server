package admin

import (
	"os"
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"
	"testing"

	"gorm.io/gorm"
)

type adminFixture struct {
	gdb          *gorm.DB
	dbConfig     *config.DBConfig
	userManageUC *UserManageUseCase
	settingsUC   *SettingsUseCase
	statUC       *StatUseCase
	userService  *service.UserService
	imageService *service.ImageService
}

func setupAdminFixture(t *testing.T) *adminFixture {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	passkeyStore := repository.NewPasskeyRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)

	dbConfig := config.NewDBConfig(settingStore)
	if err := dbConfig.InitializeSettings(); err != nil {
		t.Fatalf("InitializeSettings failed: %v", err)
	}
	dbConfig.ClearCache()

	userService := service.NewUserService(userStore, dbConfig)
	imageService := service.NewImageService(imageStore, dbConfig)
	passkeyService := service.NewPasskeyService(passkeyStore, dbConfig)
	emailService := service.NewEmailService(dbConfig)
	_ = service.NewInitService(systemStore, dbConfig)

	return &adminFixture{
		gdb:          gdb,
		dbConfig:     dbConfig,
		userManageUC: NewUserManageUseCase(userService, imageService, passkeyService),
		settingsUC:   NewSettingsUseCase(emailService),
		statUC:       NewStatUseCase(imageStore, userStore),
		userService:  userService,
		imageService: imageService,
	}
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
