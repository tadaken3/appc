package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Fatal("DefaultConfigPath() returned empty string")
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("expected config.json, got %s", filepath.Base(path))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		IssuerID:       "issuer-123",
		KeyID:          "key-456",
		PrivateKeyPath: "/path/to/key.p8",
		VendorNumber:   "12345678",
	}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.IssuerID != cfg.IssuerID {
		t.Errorf("IssuerID = %q, want %q", loaded.IssuerID, cfg.IssuerID)
	}
	if loaded.KeyID != cfg.KeyID {
		t.Errorf("KeyID = %q, want %q", loaded.KeyID, cfg.KeyID)
	}
	if loaded.PrivateKeyPath != cfg.PrivateKeyPath {
		t.Errorf("PrivateKeyPath = %q, want %q", loaded.PrivateKeyPath, cfg.PrivateKeyPath)
	}
	if loaded.VendorNumber != cfg.VendorNumber {
		t.Errorf("VendorNumber = %q, want %q", loaded.VendorNumber, cfg.VendorNumber)
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("Load() should return error for non-existent file")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	cfg := &Config{IssuerID: "test"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}
}

func TestValidate(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.p8")
	if err := os.WriteFile(keyPath, []byte("dummy"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			IssuerID:       "issuer-123",
			KeyID:          "key-456",
			PrivateKeyPath: keyPath,
			VendorNumber:   "12345678",
		}
		if err := Validate(cfg); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("missing issuer_id", func(t *testing.T) {
		cfg := &Config{KeyID: "key-456", PrivateKeyPath: keyPath, VendorNumber: "12345678"}
		if err := Validate(cfg); err == nil {
			t.Error("Validate() should return error for missing IssuerID")
		}
	})

	t.Run("missing key_id", func(t *testing.T) {
		cfg := &Config{IssuerID: "issuer-123", PrivateKeyPath: keyPath, VendorNumber: "12345678"}
		if err := Validate(cfg); err == nil {
			t.Error("Validate() should return error for missing KeyID")
		}
	})

	t.Run("missing private_key_path", func(t *testing.T) {
		cfg := &Config{IssuerID: "issuer-123", KeyID: "key-456", VendorNumber: "12345678"}
		if err := Validate(cfg); err == nil {
			t.Error("Validate() should return error for missing PrivateKeyPath")
		}
	})

	t.Run("private key file not found", func(t *testing.T) {
		cfg := &Config{
			IssuerID:       "issuer-123",
			KeyID:          "key-456",
			PrivateKeyPath: "/nonexistent/key.p8",
			VendorNumber:   "12345678",
		}
		if err := Validate(cfg); err == nil {
			t.Error("Validate() should return error when private key file does not exist")
		}
	})

	t.Run("missing vendor_number", func(t *testing.T) {
		cfg := &Config{IssuerID: "issuer-123", KeyID: "key-456", PrivateKeyPath: keyPath}
		if err := Validate(cfg); err == nil {
			t.Error("Validate() should return error for missing VendorNumber")
		}
	})
}

func TestValidateInlinePrivateKey(t *testing.T) {
	// Inline private key content should pass validation without a file.
	cfg := &Config{
		IssuerID:     "issuer-123",
		KeyID:        "key-456",
		PrivateKey:   "-----BEGIN PRIVATE KEY-----\nabc\n-----END PRIVATE KEY-----",
		VendorNumber: "12345678",
	}
	if err := Validate(cfg); err != nil {
		t.Errorf("Validate() unexpected error with inline key: %v", err)
	}
}

func TestResolveFromEnv(t *testing.T) {
	// No config file: values come entirely from the environment.
	t.Setenv(EnvIssuerID, "env-issuer")
	t.Setenv(EnvKeyID, "env-key")
	t.Setenv(EnvPrivateKey, "env-pem")
	t.Setenv(EnvVendorNumber, "env-vendor")

	cfg, err := Resolve(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if cfg.IssuerID != "env-issuer" {
		t.Errorf("IssuerID = %q, want %q", cfg.IssuerID, "env-issuer")
	}
	if cfg.KeyID != "env-key" {
		t.Errorf("KeyID = %q, want %q", cfg.KeyID, "env-key")
	}
	if cfg.PrivateKey != "env-pem" {
		t.Errorf("PrivateKey = %q, want %q", cfg.PrivateKey, "env-pem")
	}
	if cfg.VendorNumber != "env-vendor" {
		t.Errorf("VendorNumber = %q, want %q", cfg.VendorNumber, "env-vendor")
	}
}

func TestResolveEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := Save(&Config{
		IssuerID:       "file-issuer",
		KeyID:          "file-key",
		PrivateKeyPath: "/file/key.p8",
		VendorNumber:   "file-vendor",
	}, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	t.Setenv(EnvIssuerID, "env-issuer")

	cfg, err := Resolve(path)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if cfg.IssuerID != "env-issuer" {
		t.Errorf("IssuerID = %q, want env to override file", cfg.IssuerID)
	}
	if cfg.KeyID != "file-key" {
		t.Errorf("KeyID = %q, want value from file", cfg.KeyID)
	}
}

func TestPrivateKeyPEM(t *testing.T) {
	t.Run("inline content preferred", func(t *testing.T) {
		cfg := &Config{PrivateKey: "inline-pem", PrivateKeyPath: "/should/not/read"}
		data, err := cfg.PrivateKeyPEM()
		if err != nil {
			t.Fatalf("PrivateKeyPEM() error: %v", err)
		}
		if string(data) != "inline-pem" {
			t.Errorf("PrivateKeyPEM() = %q, want %q", data, "inline-pem")
		}
	})

	t.Run("reads from file path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "key.p8")
		if err := os.WriteFile(path, []byte("file-pem"), 0600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		cfg := &Config{PrivateKeyPath: path}
		data, err := cfg.PrivateKeyPEM()
		if err != nil {
			t.Fatalf("PrivateKeyPEM() error: %v", err)
		}
		if string(data) != "file-pem" {
			t.Errorf("PrivateKeyPEM() = %q, want %q", data, "file-pem")
		}
	})
}

func TestSaveFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{IssuerID: "test"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}
