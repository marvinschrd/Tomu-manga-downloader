package config_test

import (
	"path/filepath"
	"testing"

	"mangatool/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.DefaultLanguage != "en" {
		t.Errorf("expected default language 'en', got %q", cfg.DefaultLanguage)
	}
	if cfg.DefaultSource != "mangadex" {
		t.Errorf("expected default source 'mangadex', got %q", cfg.DefaultSource)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := config.DefaultConfig()
	cfg.OutputDir = "/tmp/manga"

	if err := config.SaveTo(cfg, path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	loaded, err := config.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if loaded.OutputDir != "/tmp/manga" {
		t.Errorf("expected output_dir '/tmp/manga', got %q", loaded.OutputDir)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := config.LoadFrom("/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultLanguage != "en" {
		t.Errorf("expected default 'en', got %q", cfg.DefaultLanguage)
	}
}
