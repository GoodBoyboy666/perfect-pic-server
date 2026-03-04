package middleware

import (
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var (
	testService *config.DBConfig
	testGormDB  *gorm.DB
)

func setupTestDB(t *testing.T) *gorm.DB {
	config.InitConfig("")
	gdb := testutils.SetupDB(t)
	testGormDB = gdb
	settingStore := repository.NewSettingRepository(gdb)
	testService = config.NewDBConfig(settingStore)
	testService.ClearCache()
	return gdb
}

func resetStatusCache() {
	// no-op: 状态缓存已改为依赖注入的 cache.Store，由测试自行创建与隔离。
}
