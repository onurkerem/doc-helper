package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizeSyncRoot returns an absolute path using directory entry names as
// stored on disk. On case-insensitive volumes (default APFS, typical Windows),
// this aligns spelling (e.g. "documents" vs "Documents") so config and state
// keys stay consistent with filepath.Abs and user input.
func NormalizeSyncRoot(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return resolveKnownPathCasing(abs)
}

func resolveKnownPathCasing(abs string) (string, error) {
	sep := string(filepath.Separator)
	vol := filepath.VolumeName(abs)
	rest := abs[len(vol):]
	if rest == "" {
		return vol, nil
	}
	if rest == sep {
		return vol + rest, nil
	}
	if !strings.HasPrefix(rest, sep) {
		return "", fmt.Errorf("unexpected path form: %s", abs)
	}
	parts := strings.Split(strings.Trim(rest, sep), sep)

	var cur string
	if vol != "" {
		cur = vol + sep
	} else {
		cur = sep
	}

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		entries, err := os.ReadDir(cur)
		if err != nil {
			return "", fmt.Errorf("read dir %q: %w", cur, err)
		}
		name := ""
		for _, e := range entries {
			if pathComponentMatches(e.Name(), part) {
				name = e.Name()
				break
			}
		}
		if name == "" {
			return "", fmt.Errorf("%q: no directory entry matching %q", cur, part)
		}
		cur = filepath.Join(cur, name)
	}
	return filepath.Clean(cur), nil
}

func pathComponentMatches(entryName, want string) bool {
	if entryName == want {
		return true
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return strings.EqualFold(entryName, want)
	}
	return false
}

// pathsEquivalentForSync treats two absolute paths as the same when they refer
// to the same location on case-insensitive systems. Used for config matching
// and state keys when NormalizeSyncRoot cannot run (e.g. path not yet present).
func pathsEquivalentForSync(a, b string) bool {
	ca, errA := filepath.Abs(filepath.Clean(a))
	cb, errB := filepath.Abs(filepath.Clean(b))
	if errA != nil || errB != nil {
		return false
	}
	if ca == cb {
		return true
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return strings.EqualFold(ca, cb)
	}
	return false
}
