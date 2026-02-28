package middleware

import (
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var testService *config.DBConfig

func setupTestDB(t *testing.T) *gorm.DB {
	config.InitConfig("")
	gdb := testutils.SetupDB(t)
	settingStore := repository.NewSettingRepository(gdb)
	testService = config.NewDBConfig(settingStore)
	testService.ClearCache()
	return gdb
}

func resetStatusCache() {
	statusCache.Range(func(key, value any) bool {
		statusCache.Delete(key)
		return true
	})
}
