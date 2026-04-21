package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type ScannedDir struct {
	Name    string // directory name (e.g., "subdirectory1")
	RelPath string // relative path from root (e.g., "subdirectory1")
}

type ScannedFile struct {
	RelPath     string // relative path from root (e.g., "subdirectory1/file1.md")
	Title       string // H1 extracted from file, or filename stem as fallback
	Content     string // raw markdown content
	ContentHash string // SHA-256 hex of content
	ParentDir   string // relative path of parent directory (e.g., "subdirectory1")
}

type ScanResult struct {
	Directories []ScannedDir
	Files       []ScannedFile
}

func ScanDirectory(rootPath string, dirExcludes []string, fileExcludes []string) (*ScanResult, error) {
	result := &ScanResult{}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(rootPath, path)

		if d.IsDir() {
			if path != rootPath && slices.Contains(dirExcludes, d.Name()) {
				return fs.SkipDir
			}
			if path != rootPath {
				result.Directories = append(result.Directories, ScannedDir{
					Name:    d.Name(),
					RelPath: relPath,
				})
			}
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		if slices.Contains(fileExcludes, d.Name()) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		parentDir := filepath.Dir(relPath)
		if parentDir == "." {
			parentDir = ""
		}

		title := extractTitle(string(content), d.Name())
		hash := sha256.Sum256(content)

		result.Files = append(result.Files, ScannedFile{
			RelPath:     relPath,
			Title:       title,
			Content:     string(content),
			ContentHash: fmt.Sprintf("%x", hash),
			ParentDir:   parentDir,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	slices.SortFunc(result.Directories, func(a, b ScannedDir) int {
		return strings.Compare(a.RelPath, b.RelPath)
	})
	slices.SortFunc(result.Files, func(a, b ScannedFile) int {
		return strings.Compare(a.RelPath, b.RelPath)
	})

	return result, nil
}

func extractTitle(content string, filename string) string {
	for line := range strings.SplitSeq(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	name := strings.TrimSuffix(filename, ".md")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return toTitle(name)
}

func toTitle(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
