package handler

import (
	"os"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/testutils"
)

// 测试内容：为 handler 包测试初始化配置环境并在结束时清理。
func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "perfect-pic-handler-config-*")
	if err != nil {
		panic(err)
	}

	envs := []testutils.SavedEnv{
		testutils.SetEnv("PERFECT_PIC_SERVER_MODE", "debug"),
		testutils.SetEnv("PERFECT_PIC_JWT_SECRET", "test_secret"),
		testutils.SetEnv("PERFECT_PIC_JWT_EXPIRATION_HOURS", "24"),
		testutils.SetEnv("PERFECT_PIC_REDIS_ENABLED", "false"),
	}
	config.InitConfig(tmpDir)

	code := m.Run()

	testutils.RestoreEnv(envs)
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}
