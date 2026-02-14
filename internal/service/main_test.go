package service

import (
	"os"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/testutils"
)

// 测试内容：为 service 包测试初始化配置环境并在结束时清理。
func TestMain(m *testing.M) {
	// 为依赖配置的测试提供稳定默认值（JWT 过期时间、上传前缀等）。
	tmpDir, err := os.MkdirTemp("", "perfect-pic-config-*")
	if err != nil {
		panic(err)
	}

	envs := []testutils.SavedEnv{
		testutils.SetEnv("PERFECT_PIC_SERVER_MODE", "debug"),
		testutils.SetEnv("PERFECT_PIC_JWT_SECRET", "test_secret"),
		testutils.SetEnv("PERFECT_PIC_REDIS_ENABLED", "false"),
	}
	config.InitConfigWithoutWatch(tmpDir)

	code := m.Run()

	testutils.RestoreEnv(envs)
	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}
