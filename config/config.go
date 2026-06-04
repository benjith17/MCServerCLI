package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Name        string `toml:"name"`
	Directory   string `toml:"directory"`
	StartScript string `toml:"start_script"`
	Version     string `toml:"version"`
	Autostart   bool   `toml:"autostart"`
}

type Config struct {
	Servers []ServerConfig `toml:"servers"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaultConfigPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "mcmanager", "config.toml"), nil
}
