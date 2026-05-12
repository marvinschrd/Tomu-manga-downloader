package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	OutputDir       string `toml:"output_dir"`
	DefaultLanguage string `toml:"default_language"`
	DefaultSource   string `toml:"default_source"`
}

func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		OutputDir:       filepath.Join(home, "manga"),
		DefaultLanguage: "en",
		DefaultSource:   "mangadex",
	}
}

func Load() (Config, error) {
	return LoadFrom(defaultPath())
}

func Save(cfg Config) error {
	return SaveTo(cfg, defaultPath())
}

func LoadFrom(path string) (Config, error) {
	cfg := DefaultConfig()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}

func SaveTo(cfg Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func defaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mangatool", "config.toml")
}
