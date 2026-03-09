package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigWhenFileMissingReturnsEmptyString(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got := loadConfig()
	if got != "" {
		t.Fatalf("expected empty string when config missing, got %q", got)
	}
}

func TestSaveAndLoadConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	saveConfig("abc123")

	got := loadConfig()
	if got != "abc123" {
		t.Fatalf("expected API key %q, got %q", "abc123", got)
	}

	// Ensure the file was written where we expect.
	path := filepath.Join(dir, ".cloud", "config.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file at %s: %v", path, err)
	}
}

func TestLoadConfigWhenInvalidJSONReturnsEmptyString(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := filepath.Join(dir, ".cloud", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := loadConfig()
	if got != "" {
		t.Fatalf("expected empty string on invalid json, got %q", got)
	}
}
