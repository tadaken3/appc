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
