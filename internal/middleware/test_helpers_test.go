package middleware

import (
	"fmt"
	"sync/atomic"
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var testDBSeq int64

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	seq := atomic.AddInt64(&testDBSeq, 1)
	dsn := fmt.Sprintf("file:ppmw_%d?mode=memory&cache=shared", seq)
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := gdb.AutoMigrate(&model.User{}, &model.Setting{}, &model.Image{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	db.DB = gdb
	service.ClearCache()
	return gdb
}

func resetStatusCache() {
	statusCache.Range(func(key, value any) bool {
		statusCache.Delete(key)
		return true
	})
}
