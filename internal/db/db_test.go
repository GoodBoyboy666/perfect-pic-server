package db

import (
	"os"
	"path/filepath"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/model"
)

func TestInitDB_SQLiteTempFile(t *testing.T) {
	tmp := t.TempDir()
	cfgDir := filepath.Join(tmp, "cfg")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("mkdir cfg: %v", err)
	}

	dbFile := filepath.Join(tmp, "db", "test.db")
	t.Setenv("PERFECT_PIC_SERVER_MODE", "debug")
	t.Setenv("PERFECT_PIC_DATABASE_TYPE", "sqlite")
	t.Setenv("PERFECT_PIC_DATABASE_FILENAME", dbFile)

	config.InitConfig(cfgDir)
	InitDB()

	if DB == nil {
		t.Fatalf("expected DB to be initialized")
	}
	if !DB.Migrator().HasTable(&model.User{}) {
		t.Fatalf("expected users table to exist")
	}
	if !DB.Migrator().HasTable(&model.Setting{}) {
		t.Fatalf("expected settings table to exist")
	}
	if !DB.Migrator().HasTable(&model.Image{}) {
		t.Fatalf("expected images table to exist")
	}

	sqlDB, err := DB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
