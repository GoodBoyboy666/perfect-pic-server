package admin

import (
	"os"
	"testing"

	"perfect-pic-server/internal/config"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "perfect-pic-admin-handler-config-*")
	if err != nil {
		panic(err)
	}

	_ = os.Setenv("PERFECT_PIC_SERVER_MODE", "debug")
	_ = os.Setenv("PERFECT_PIC_JWT_SECRET", "test_secret")
	_ = os.Setenv("PERFECT_PIC_JWT_EXPIRATION_HOURS", "24")
	_ = os.Setenv("PERFECT_PIC_REDIS_ENABLED", "false")
	config.InitConfig(tmpDir)

	code := m.Run()

	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}
