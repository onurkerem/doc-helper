package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PageState struct {
	PageID       string `json:"page_id"`
	Title        string `json:"title"`
	Version      int    `json:"version"`
	ContentHash  string `json:"content_hash"`
	ParentPageID string `json:"parent_page_id,omitempty"`
}

type SyncState map[string]PageState // key: relative path
type StateData map[string]SyncState // key: absolute root path

func LoadState() (StateData, error) {
	data, err := os.ReadFile(StatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return StateData{}, nil
		}
		return nil, fmt.Errorf("reading state: %w", err)
	}

	if len(data) == 0 {
		return StateData{}, nil
	}

	var state StateData
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: corrupted state file, starting fresh: %v\n", err)
		return StateData{}, nil
	}

	return state, nil
}

func SaveState(state StateData) error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(StatePath(), data, 0600); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}

	return nil
}

func (sd StateData) GetPageState(rootPath, relPath string) *PageState {
	syncState, ok := sd[rootPath]
	if !ok {
		return nil
	}
	ps, ok := syncState[relPath]
	if !ok {
		return nil
	}
	return &ps
}

func (sd StateData) SetPageState(rootPath, relPath string, ps PageState) {
	if sd[rootPath] == nil {
		sd[rootPath] = SyncState{}
	}
	sd[rootPath][relPath] = ps
}
