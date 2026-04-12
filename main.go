package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: doc-helper <path>\n")
		os.Exit(1)
	}

	root := os.Args[1]

	info, err := os.Stat(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", root)
		os.Exit(1)
	}

	var mdFiles []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	if len(mdFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No markdown files found in %s\n", root)
		os.Exit(1)
	}

	slices.Sort(mdFiles)

	var sb strings.Builder
	for i, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		if i > 0 {
			sb.WriteString("\n")
		}
		rel, _ := filepath.Rel(".", path)
		fmt.Fprintf(&sb, "<!-- %s -->\n", rel)
		sb.Write(content)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			sb.WriteString("\n")
		}
	}

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(sb.String())
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Copied %d markdown file(s) to clipboard.\n", len(mdFiles))
}
