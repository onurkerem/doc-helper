package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SyncConfig struct {
	Path             string `json:"path"`
	ConfluenceBaseURL string `json:"confluence_base_url"`
	Email            string `json:"email"`
	APIToken         string `json:"api_token"`
	ParentPageID     string `json:"parent_page_id"`
}

type Config struct {
	Syncs []SyncConfig `json:"syncs"`
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".doc-helper")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func StatePath() string {
	return filepath.Join(ConfigDir(), "state.json")
}

func ensureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0700)
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) FindSync(absPath string) *SyncConfig {
	canon, errAbs := NormalizeSyncRoot(absPath)
	for i := range c.Syncs {
		cfgPath := c.Syncs[i].Path
		cfgCanon, errCfg := NormalizeSyncRoot(cfgPath)
		if errAbs == nil && errCfg == nil && cfgCanon == canon {
			return &c.Syncs[i]
		}
		if pathsEquivalentForSync(cfgPath, absPath) {
			return &c.Syncs[i]
		}
	}
	return nil
}
