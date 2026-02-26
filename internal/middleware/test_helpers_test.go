package middleware

import (
	"testing"

	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	"perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *service.AppService

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	settingStore := settingsrepo.NewSettingRepository(gdb)
	testService = service.NewAppService(settingStore)
	testService.ClearCache()
	return gdb
}

func resetStatusCache() {
	statusCache.Range(func(key, value any) bool {
		statusCache.Delete(key)
		return true
	})
}
