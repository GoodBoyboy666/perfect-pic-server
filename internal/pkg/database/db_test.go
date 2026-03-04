package database

import (
	"path/filepath"
	"testing"

	"perfect-pic-server/internal/model"
)

// 测试内容：验证使用 sqlite 临时文件初始化数据库并创建核心表。
func TestNewGormDB_SQLiteTempFile(t *testing.T) {
	tmp := t.TempDir()
	dbFile := filepath.Join(tmp, "db", "test.db")

	gdb, err := NewGormDB(&DbConnectionConfig{
		Type:     "sqlite",
		Filename: dbFile,
	})
	if err != nil {
		t.Fatalf("NewGormDB failed: %v", err)
	}

	if gdb == nil {
		t.Fatalf("期望 DB to be initialized")
	}
	if !gdb.Migrator().HasTable(&model.User{}) {
		t.Fatalf("期望 users table to exist")
	}
	if !gdb.Migrator().HasTable(&model.Setting{}) {
		t.Fatalf("期望 settings table to exist")
	}
	if !gdb.Migrator().HasTable(&model.Image{}) {
		t.Fatalf("期望 images table to exist")
	}

	sqlDB, err := gdb.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
