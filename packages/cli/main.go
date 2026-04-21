package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var root string
	var excludes []string
	var confluence bool
	var dryRun bool
	var force bool

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--exclude":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --exclude requires a value\n")
				os.Exit(1)
			}
			i++
			excludes = append(excludes, strings.Split(args[i], ",")...)
		case "--confluence":
			confluence = true
		case "--dry-run":
			dryRun = true
		case "--force":
			force = true
		default:
			if root != "" {
				fmt.Fprintf(os.Stderr, "Error: unexpected argument %s\n", args[i])
				os.Exit(1)
			}
			root = args[i]
		}
	}

	if root == "" {
		fmt.Fprintf(os.Stderr, "Usage: doc-helper <path> [--exclude <dir>[,<dir>...]] [--confluence] [--dry-run] [--force]\n")
		os.Exit(1)
	}

	if force && !confluence {
		fmt.Fprintf(os.Stderr, "Error: --force is only valid with --confluence\n")
		os.Exit(1)
	}

	info, err := os.Stat(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", root)
		os.Exit(1)
	}

	result, err := ScanDirectory(root, excludes, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(result.Files) == 0 {
		fmt.Fprintf(os.Stderr, "No markdown files found in %s\n", root)
		os.Exit(1)
	}

	// Clipboard (existing behavior)
	copyToClipboard(root, result)

	// Confluence sync
	if confluence {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if canon, err := NormalizeSyncRoot(root); err == nil {
			absRoot = canon
		}
		if err := runConfluenceSync(absRoot, excludes, dryRun, force); err != nil {
			fmt.Fprintf(os.Stderr, "Confluence sync error: %v\n", err)
			os.Exit(1)
		}
	}
}

func copyToClipboard(root string, result *ScanResult) {
	var sb strings.Builder
	for i, file := range result.Files {
		fullPath := filepath.Join(root, file.RelPath)
		if i > 0 {
			sb.WriteString("\n")
		}
		rel, _ := filepath.Rel(".", fullPath)
		fmt.Fprintf(&sb, "<!-- %s -->\n", rel)
		sb.WriteString(file.Content)
		if len(file.Content) > 0 && file.Content[len(file.Content)-1] != '\n' {
			sb.WriteString("\n")
		}
	}

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(sb.String())
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Copied %d markdown file(s) to clipboard.\n", len(result.Files))
}

func runConfluenceSync(absRoot string, excludes []string, dryRun, force bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	syncCfg := cfg.FindSync(absRoot)
	if syncCfg == nil {
		return fmt.Errorf("path %s is not configured for Confluence sync. Add it to %s", absRoot, ConfigPath())
	}

	return RunSync(syncCfg, absRoot, excludes, syncCfg.ExcludeFiles, dryRun, force)
}
