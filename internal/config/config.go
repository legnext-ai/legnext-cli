package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultAPIBase = "https://api.legnext.ai"
	EnvAPIKey      = "LEGNEXT_API_KEY"
	EnvAPIBase     = "LEGNEXT_API_BASE"
)

type Config struct {
	APIKey  string `json:"api_key,omitempty"`
	APIBase string `json:"api_base,omitempty"`
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".legnext", "config.json"), nil
}

// Load reads the on-disk config, returning an empty Config if absent.
// It does NOT apply env overrides — call Resolve for the effective values.
func Load() (*Config, error) {
	cfg := &Config{}
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// Resolve returns the effective (apiKey, apiBase), applying env overrides on
// top of the on-disk config. Env wins over file; DefaultAPIBase fills empty.
func Resolve() (apiKey, apiBase string, err error) {
	cfg, err := Load()
	if err != nil {
		return "", "", err
	}
	apiKey = cfg.APIKey
	if v := os.Getenv(EnvAPIKey); v != "" {
		apiKey = v
	}
	apiBase = cfg.APIBase
	if v := os.Getenv(EnvAPIBase); v != "" {
		apiBase = v
	}
	if apiBase == "" {
		apiBase = DefaultAPIBase
	}
	return apiKey, apiBase, nil
}

func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// MaskKey returns e.g. "af3e...1a16" for safe display in logs.
func MaskKey(key string) string {
	if len(key) < 12 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
