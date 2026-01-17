package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gopac-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// No config file
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error when no file exists: %v", err)
	}
	if cfg.AURHelper != "" {
		t.Errorf("Expected empty AURHelper, got %q", cfg.AURHelper)
	}

	// With config file
	configDir := filepath.Join(tmpDir, "gopac")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	content := []byte("aur_helper: yay")
	err = os.WriteFile(configFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() returned error with valid file: %v", err)
	}
	if cfg.AURHelper != "yay" {
		t.Errorf("Expected AURHelper 'yay', got %q", cfg.AURHelper)
	}
}
