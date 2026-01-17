package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AURHelper string `yaml:"aur_helper"`
	Theme     string `yaml:"theme"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return cfg, err
	}

	configPath := filepath.Join(configDir, "gopac", "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
