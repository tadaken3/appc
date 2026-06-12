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

	// PrivateKey holds the PEM-encoded private key content directly.
	// It is populated only from the environment (APPC_PRIVATE_KEY) and is
	// never written to the config file. Useful in cloud environments where
	// placing a .p8 file on disk is inconvenient.
	PrivateKey string `json:"-"`
}

// Environment variable names. When set, they override values loaded from the
// config file. This lets the tool run in cloud environments (e.g. Claude Code
// on the web) without an on-disk config file.
const (
	EnvIssuerID       = "APPC_ISSUER_ID"
	EnvKeyID          = "APPC_KEY_ID"
	EnvPrivateKeyPath = "APPC_PRIVATE_KEY_PATH"
	EnvPrivateKey     = "APPC_PRIVATE_KEY"
	EnvVendorNumber   = "APPC_VENDOR_NUMBER"
)

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "appc", "config.json")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.PrivateKeyPath = expandHome(cfg.PrivateKeyPath)
	return &cfg, nil
}

// Resolve loads the config from path (if it exists) and overlays any values
// set via environment variables. A missing config file is not an error as long
// as the required values are supplied through the environment.
func Resolve(path string) (*Config, error) {
	cfg := &Config{}
	if path != "" {
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config: %w", err)
			}
		case errors.Is(err, fs.ErrNotExist):
			// No config file: rely entirely on environment variables.
		default:
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	applyEnv(cfg)
	cfg.PrivateKeyPath = expandHome(cfg.PrivateKeyPath)
	return cfg, nil
}

// applyEnv overrides cfg fields with environment variables when they are set.
func applyEnv(cfg *Config) {
	if v := os.Getenv(EnvIssuerID); v != "" {
		cfg.IssuerID = v
	}
	if v := os.Getenv(EnvKeyID); v != "" {
		cfg.KeyID = v
	}
	if v := os.Getenv(EnvPrivateKeyPath); v != "" {
		cfg.PrivateKeyPath = v
	}
	if v := os.Getenv(EnvPrivateKey); v != "" {
		cfg.PrivateKey = v
	}
	if v := os.Getenv(EnvVendorNumber); v != "" {
		cfg.VendorNumber = v
	}
}

// PrivateKeyPEM returns the PEM-encoded private key, preferring inline content
// (from APPC_PRIVATE_KEY) over the file at PrivateKeyPath.
func (c *Config) PrivateKeyPEM() ([]byte, error) {
	if c.PrivateKey != "" {
		return []byte(c.PrivateKey), nil
	}
	data, err := os.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}
	return data, nil
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
	// The private key may be supplied inline (APPC_PRIVATE_KEY) or via a file
	// path. Inline content takes precedence and needs no file check.
	if cfg.PrivateKey == "" {
		if cfg.PrivateKeyPath == "" {
			return fmt.Errorf("private key is not set: run 'appc configure' or set %s / %s", EnvPrivateKey, EnvPrivateKeyPath)
		}
		if _, err := os.Stat(cfg.PrivateKeyPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("private key file not found")
			}
			return fmt.Errorf("checking private key file: %w", err)
		}
	}
	if cfg.VendorNumber == "" {
		return fmt.Errorf("vendor_number is not set: run 'appc configure'")
	}
	return nil
}

// ValidateFromPath resolves the config from path (overlaying environment
// variables) and validates it.
func ValidateFromPath(path string) error {
	cfg, err := Resolve(path)
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
