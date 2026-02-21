package middleware

import (
	"testing"

	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *service.AppService

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	testService = service.NewAppService(repository.NewRepositories(
		repository.NewUserRepository(gdb),
		repository.NewImageRepository(gdb),
		repository.NewSettingRepository(gdb),
		repository.NewSystemRepository(gdb),
	))
	testService.ClearCache()
	return gdb
}

func resetStatusCache() {
	statusCache.Range(func(key, value any) bool {
		statusCache.Delete(key)
		return true
	})
}
