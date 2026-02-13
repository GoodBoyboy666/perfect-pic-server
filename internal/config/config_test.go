package config

import (
	"os"
	"testing"
)

func TestInitConfig_SetsDefaults(t *testing.T) {
	dir := t.TempDir()

	// Ensure we're not in release mode (release mode can fatal if secret is unsafe).
	t.Setenv("PERFECT_PIC_SERVER_MODE", "debug")
	t.Setenv("PERFECT_PIC_JWT_SECRET", "")

	InitConfig(dir)

	cfg := Get()
	if cfg.Server.Port == "" {
		t.Fatalf("expected default server.port to be set")
	}
	if cfg.JWT.Secret == "" {
		t.Fatalf("expected JWT secret to be set in non-release mode")
	}
	if GetConfigDir() != dir {
		t.Fatalf("expected config dir %q, got %q", dir, GetConfigDir())
	}

	// Touch a config file name to ensure the directory is writable (sanity check for tests).
	if err := os.WriteFile(dir+string(os.PathSeparator)+"_test_write", []byte("ok"), 0644); err != nil {
		t.Fatalf("expected temp config dir to be writable: %v", err)
	}
}
