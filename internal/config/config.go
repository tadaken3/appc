package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	IssuerID       string `json:"issuer_id"`
	KeyID          string `json:"key_id"`
	PrivateKeyPath string `json:"private_key_path"`
	VendorNumber   string `json:"vendor_number"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "appc", "config.json")
}

func Load(path string) (*Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		if !hasEnvConfig() {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}
	applyEnvOverrides(&cfg)
	cfg.PrivateKeyPath = expandHome(cfg.PrivateKeyPath)
	return &cfg, nil
}

func hasEnvConfig() bool {
	return os.Getenv("APPC_ISSUER_ID") != "" ||
		os.Getenv("APPC_KEY_ID") != "" ||
		os.Getenv("APPC_PRIVATE_KEY_PATH") != "" ||
		os.Getenv("APPC_VENDOR_NUMBER") != ""
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("APPC_ISSUER_ID"); v != "" {
		cfg.IssuerID = v
	}
	if v := os.Getenv("APPC_KEY_ID"); v != "" {
		cfg.KeyID = v
	}
	if v := os.Getenv("APPC_PRIVATE_KEY_PATH"); v != "" {
		cfg.PrivateKeyPath = v
	}
	if v := os.Getenv("APPC_VENDOR_NUMBER"); v != "" {
		cfg.VendorNumber = v
	}
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.IssuerID == "" {
		return fmt.Errorf("issuer_id is not set: run 'appc configure'")
	}
	if cfg.KeyID == "" {
		return fmt.Errorf("key_id is not set: run 'appc configure'")
	}
	if cfg.PrivateKeyPath == "" {
		return fmt.Errorf("private_key_path is not set: run 'appc configure'")
	}
	if _, err := os.Stat(cfg.PrivateKeyPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("private key file not found")
		}
		return fmt.Errorf("checking private key file: %w", err)
	}
	if cfg.VendorNumber == "" {
		return fmt.Errorf("vendor_number is not set: run 'appc configure'")
	}
	return nil
}

// ValidateFromPath loads the config from path and validates it.
func ValidateFromPath(path string) error {
	cfg, err := Load(path)
	if err != nil {
		return fmt.Errorf("loading config: %w (run 'appc configure' to set up credentials)", err)
	}
	return Validate(cfg)
}

func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
