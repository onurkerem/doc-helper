package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory_NoExcludes(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Readme"), 0644)
	os.WriteFile(filepath.Join(root, "OTHER.md"), []byte("# Other"), 0644)

	result, err := ScanDirectory(root, nil, nil)
	if err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
}

func TestScanDirectory_FileExcludes(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Readme"), 0644)
	os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# Claude"), 0644)

	result, err := ScanDirectory(root, nil, []string{"CLAUDE.md"})
	if err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if result.Files[0].RelPath != "README.md" {
		t.Fatalf("expected README.md, got %s", result.Files[0].RelPath)
	}
}

func TestScanDirectory_FileExcludesInSubdirectory(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "sub"), 0755)
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Readme"), 0644)
	os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# Claude"), 0644)
	os.WriteFile(filepath.Join(root, "sub", "CLAUDE.md"), []byte("# Sub Claude"), 0644)
	os.WriteFile(filepath.Join(root, "sub", "guide.md"), []byte("# Guide"), 0644)

	result, err := ScanDirectory(root, nil, []string{"CLAUDE.md"})
	if err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(result.Files), fileNames(result.Files))
	}
	expected := map[string]bool{"README.md": true, filepath.Join("sub", "guide.md"): true}
	for _, f := range result.Files {
		if !expected[f.RelPath] {
			t.Errorf("unexpected file: %s", f.RelPath)
		}
	}
}

func TestScanDirectory_DirAndFileExcludes(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "node_modules"), 0755)
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Readme"), 0644)
	os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# Claude"), 0644)
	os.WriteFile(filepath.Join(root, "node_modules", "pkg.md"), []byte("# Pkg"), 0644)

	result, err := ScanDirectory(root, []string{"node_modules"}, []string{"CLAUDE.md"})
	if err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
	if result.Files[0].RelPath != "README.md" {
		t.Fatalf("expected README.md, got %s", result.Files[0].RelPath)
	}
}

func TestScanDirectory_EmptyFileExcludes(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# Readme"), 0644)

	result, err := ScanDirectory(root, nil, []string{})
	if err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
}

func fileNames(files []ScannedFile) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.RelPath
	}
	return names
}
