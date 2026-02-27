package middleware

import (
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *service.Service

func setupTestDB(t *testing.T) *gorm.DB {
	config.InitConfig("")
	gdb := testutils.SetupDB(t)
	userStore := repository.NewUserRepository(gdb)
	imageStore := repository.NewImageRepository(gdb)
	settingStore := repository.NewSettingRepository(gdb)
	systemStore := repository.NewSystemRepository(gdb)
	testService = service.NewAppService(userStore, imageStore, settingStore, systemStore)
	testService.ClearCache()
	return gdb
}

func resetStatusCache() {
	statusCache.Range(func(key, value any) bool {
		statusCache.Delete(key)
		return true
	})
}
