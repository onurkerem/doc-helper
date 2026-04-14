package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunSync(cfg *SyncConfig, rootPath string, excludes []string, dryRun bool) error {
	result, err := ScanDirectory(rootPath, excludes)
	if err != nil {
		return fmt.Errorf("scanning directory: %w", err)
	}

	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	client := NewConfluenceClient(cfg.ConfluenceBaseURL, cfg.Email, cfg.APIToken)
	converter := NewMarkdownConverter()

	// Auto-resolve space_id from parent page
	parentPage, err := client.GetPage(cfg.ParentPageID)
	if err != nil {
		return fmt.Errorf("fetching parent page %s: %w", cfg.ParentPageID, err)
	}
	if parentPage == nil {
		return fmt.Errorf("parent page %s not found", cfg.ParentPageID)
	}
	spaceID := parentPage.SpaceID

	fmt.Printf("Confluence sync: %s → space %s, parent page %s\n", rootPath, spaceID, cfg.ParentPageID)

	created := 0
	updated := 0
	skipped := 0

	// Helper to find parent Confluence page ID for a given relative directory path
	getParentPageID := func(relDirPath string) string {
		if relDirPath == "" {
			return cfg.ParentPageID
		}
		ps := state.GetPageState(rootPath, relDirPath)
		if ps != nil {
			return ps.PageID
		}
		return cfg.ParentPageID
	}

	// Helper to find existing page by title under a parent (fallback for lost state)
	findPageByTitle := func(parentID, title string) *ConfluencePage {
		children, err := client.GetChildPages(parentID)
		if err != nil {
			return nil
		}
		for _, child := range children {
			if child.Title == title {
				return &child
			}
		}
		return nil
	}

	// Process directories first (create hierarchy pages)
	for _, dir := range result.Directories {
		existing := state.GetPageState(rootPath, dir.RelPath)

		if existing != nil {
			skipped++
			continue
		}

		// Determine parent page for this directory
		parentDir := filepath.Dir(dir.RelPath)
		if parentDir == "." {
			parentDir = ""
		}
		parentID := getParentPageID(parentDir)

		if dryRun {
			fmt.Printf("  [dry-run] Would create page: %s (under %s)\n", dir.Name, parentID)
			continue
		}

		// Check for existing page with same title (lost state)
		existingPage := findPageByTitle(parentID, dir.Name)

		if existingPage != nil {
			// Page exists but we lost state — recover it
			state.SetPageState(rootPath, dir.RelPath, PageState{
				PageID:       existingPage.ID,
				Title:        existingPage.Title,
				Version:      existingPage.Version.Number,
				ContentHash:  "",
				ParentPageID: parentID,
			})
			skipped++
			fmt.Printf("  Recovered state for: %s (page %s)\n", dir.Name, existingPage.ID)
			continue
		}

		// Create new empty directory page
		page, err := client.CreatePage(spaceID, dir.Name, "<p></p>", parentID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error creating page for %s: %v\n", dir.RelPath, err)
			continue
		}

		state.SetPageState(rootPath, dir.RelPath, PageState{
			PageID:       page.ID,
			Title:        dir.Name,
			Version:      1,
			ContentHash:  "",
			ParentPageID: parentID,
		})
		created++
		fmt.Printf("  Created: %s\n", dir.Name)
	}

	// Save state after directories so files can reference them
	if !dryRun {
		if err := SaveState(state); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save state after directories: %v\n", err)
		}
	}

	// Process files
	for _, file := range result.Files {
		existing := state.GetPageState(rootPath, file.RelPath)

		// Check if content unchanged
		if existing != nil && existing.ContentHash == file.ContentHash {
			skipped++
			continue
		}

		// Find parent page ID
		parentID := getParentPageID(file.ParentDir)

		// Convert markdown to HTML
		html, err := converter.Convert(file.Content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error converting %s: %v\n", file.RelPath, err)
			continue
		}

		title := file.Title
		if title == "" {
			title = "Untitled"
		}

		if dryRun {
			if existing != nil {
				fmt.Printf("  [dry-run] Would update page: %s\n", title)
			} else {
				fmt.Printf("  [dry-run] Would create page: %s (under %s)\n", title, parentID)
			}
			continue
		}

		if existing != nil {
			// Update existing page
			page, err := client.UpdatePage(existing.PageID, title, html, existing.Version)
			if err != nil {
				// Version conflict — fetch current version and retry
				if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "Conflict") {
					currentPage, fetchErr := client.GetPage(existing.PageID)
					if fetchErr == nil && currentPage != nil {
						page, err = client.UpdatePage(existing.PageID, title, html, currentPage.Version.Number)
					}
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Error updating %s: %v\n", file.RelPath, err)
					continue
				}
			}

			state.SetPageState(rootPath, file.RelPath, PageState{
				PageID:       page.ID,
				Title:        title,
				Version:      page.Version.Number,
				ContentHash:  file.ContentHash,
				ParentPageID: parentID,
			})
			updated++
			fmt.Printf("  Updated: %s\n", title)
		} else {
			// Check for existing page with same title (lost state)
			existingPage := findPageByTitle(parentID, title)

			if existingPage != nil {
				// Update the existing page
				page, err := client.UpdatePage(existingPage.ID, title, html, existingPage.Version.Number)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Error updating existing %s: %v\n", file.RelPath, err)
					continue
				}

				state.SetPageState(rootPath, file.RelPath, PageState{
					PageID:       page.ID,
					Title:        title,
					Version:      page.Version.Number,
					ContentHash:  file.ContentHash,
					ParentPageID: parentID,
				})
				updated++
				fmt.Printf("  Updated (recovered): %s\n", title)
				continue
			}

			// Create new page
			page, err := client.CreatePage(spaceID, title, html, parentID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error creating page for %s: %v\n", file.RelPath, err)
				continue
			}

			state.SetPageState(rootPath, file.RelPath, PageState{
				PageID:       page.ID,
				Title:        title,
				Version:      1,
				ContentHash:  file.ContentHash,
				ParentPageID: parentID,
			})
			created++
			fmt.Printf("  Created: %s\n", title)
		}

		// Save state incrementally
		if err := SaveState(state); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
		}
	}

	// Final state save
	if !dryRun {
		if err := SaveState(state); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
	}

	action := "Synced"
	if dryRun {
		action = "Would sync"
	}
	fmt.Printf("%s: %d created, %d updated, %d skipped\n", action, created, updated, skipped)

	return nil
}
