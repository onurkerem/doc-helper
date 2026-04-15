package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPathsEquivalentForSync(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("case folding only on darwin/windows")
	}
	a := filepath.Join(string(filepath.Separator), "tmp", "ABC", "x")
	b := filepath.Join(string(filepath.Separator), "tmp", "abc", "x")
	if !pathsEquivalentForSync(a, b) {
		t.Fatalf("expected equivalent paths")
	}
}

func TestNormalizeSyncRoot_caseInsensitiveVolume(t *testing.T) {
	if runtime.GOOS == "linux" {
		// Typical Linux filesystems are case-sensitive; wrong-case paths do not resolve.
		t.Skip("case-insensitive path test")
	}
	tmp := t.TempDir()
	realName := filepath.Join(tmp, "RealNameDir")
	if err := os.Mkdir(realName, 0o755); err != nil {
		t.Fatal(err)
	}
	// Wrong-case path that still opens on darwin/windows:
	wrongCase := filepath.Join(tmp, "realnamedir")
	if _, err := os.Stat(wrongCase); err != nil {
		t.Skip("filesystem appears case-sensitive:", err)
	}
	got, err := NormalizeSyncRoot(wrongCase)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(realName)
	if err != nil {
		want = realName
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
