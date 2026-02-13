package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSecureJoin_AllowsWithinBase(t *testing.T) {
	base := t.TempDir()

	got, err := SecureJoin(base, filepath.Join("a", "b", "c.txt"))
	if err != nil {
		t.Fatalf("SecureJoin returned error: %v", err)
	}

	baseAbs, _ := filepath.Abs(base)
	if !strings.HasPrefix(strings.ToLower(got), strings.ToLower(baseAbs+string(os.PathSeparator))) && !strings.EqualFold(got, baseAbs) {
		t.Fatalf("expected joined path to be under base, got=%q base=%q", got, baseAbs)
	}
}

func TestSecureJoin_RejectsAbsoluteInput(t *testing.T) {
	base := t.TempDir()
	abs := filepath.Join(base, "x.txt")

	_, err := SecureJoin(base, abs)
	if err == nil {
		t.Fatalf("expected error for absolute input path")
	}
}

func TestSecureJoin_RejectsTraversalOutsideBase(t *testing.T) {
	base := t.TempDir()
	_, err := SecureJoin(base, filepath.Join("..", "escape.txt"))
	if err == nil {
		t.Fatalf("expected error for traversal")
	}
}

func TestEnsurePathNotSymlink_NonExistentOK(t *testing.T) {
	p := filepath.Join(t.TempDir(), "does-not-exist")
	if err := EnsurePathNotSymlink(p); err != nil {
		t.Fatalf("expected nil for non-existent path, got: %v", err)
	}
}

func TestEnsureNoSymlinkBetween_RejectsOutsideBase(t *testing.T) {
	base := t.TempDir()
	outside := t.TempDir()

	if err := EnsureNoSymlinkBetween(base, outside); err == nil {
		t.Fatalf("expected error when target is outside base")
	}
}

func TestEnsureNoSymlinkBetween_RejectsCrossVolumeOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific")
	}

	base := t.TempDir()
	// Any different drive letter is considered a different volume even if it doesn't exist.
	target := `Z:\somewhere`
	if err := EnsureNoSymlinkBetween(base, target); err == nil {
		t.Fatalf("expected cross-volume error")
	}
}
