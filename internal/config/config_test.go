package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toggl-tui", "config.toml")

	want := &Config{
		APIToken:      "test-token-abc123",
		WorkspaceID:   42,
		WorkspaceName: "My Workspace",
	}

	if err := SaveTo(want, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got == nil {
		t.Fatal("LoadFrom returned nil config")
	}
	if got.APIToken != want.APIToken {
		t.Errorf("APIToken = %q, want %q", got.APIToken, want.APIToken)
	}
	if got.WorkspaceID != want.WorkspaceID {
		t.Errorf("WorkspaceID = %d, want %d", got.WorkspaceID, want.WorkspaceID)
	}
	if got.WorkspaceName != want.WorkspaceName {
		t.Errorf("WorkspaceName = %q, want %q", got.WorkspaceName, want.WorkspaceName)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "config.toml")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom should not error for missing file: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config for missing file, got %+v", cfg)
	}
}

func TestLoadMalformedTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(path, []byte("this is {not valid toml!!!"), 0o600)

	cfg, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML")
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got %+v", cfg)
	}
}

func TestSaveCreatesDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "config.toml")

	err := SaveTo(&Config{APIToken: "tok"}, path)
	if err != nil {
		t.Fatalf("SaveTo should create intermediate dirs: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestSaveFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	SaveTo(&Config{APIToken: "secret"}, path)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 600", perm)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	SaveTo(&Config{APIToken: "old-token", WorkspaceID: 1}, path)
	SaveTo(&Config{APIToken: "new-token", WorkspaceID: 2}, path)

	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got.APIToken != "new-token" {
		t.Errorf("APIToken = %q, want 'new-token'", got.APIToken)
	}
	if got.WorkspaceID != 2 {
		t.Errorf("WorkspaceID = %d, want 2", got.WorkspaceID)
	}
}
