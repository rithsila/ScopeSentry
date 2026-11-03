package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultServerURL = "http://localhost:8082"
	DefaultTimeout   = 30 * time.Second
)

type Config struct {
	ServerURL string        `yaml:"server_url" json:"server_url"`
	APIKey    string        `yaml:"api_key" json:"api_key"`
	Timeout   time.Duration `yaml:"timeout" json:"timeout"`
}

type fileConfig struct {
	ServerURL string `yaml:"server_url"`
	APIKey    string `yaml:"api_key"`
	Timeout   string `yaml:"timeout"`
}

func LoadConfig(flagURL, flagAPIKey string) *Config {
	cfg := &Config{
		ServerURL: DefaultServerURL,
		Timeout:   DefaultTimeout,
	}

	// 1. Try loading from file
	if fCfg, err := loadConfigFile(); err == nil && fCfg != nil {
		if fCfg.ServerURL != "" {
			cfg.ServerURL = fCfg.ServerURL
		}
		if fCfg.APIKey != "" {
			cfg.APIKey = fCfg.APIKey
		}
		if fCfg.Timeout != "" {
			if d, err := time.ParseDuration(fCfg.Timeout); err == nil {
				cfg.Timeout = d
			}
		}
	}

	// 2. Env vars override file
	if envURL := strings.TrimSpace(os.Getenv("SCOPESENTRY_URL")); envURL != "" {
		cfg.ServerURL = envURL
	}
	if envKey := strings.TrimSpace(os.Getenv("SCOPESENTRY_API_KEY")); envKey != "" {
		cfg.APIKey = envKey
	}

	// 3. Flags override env vars
	if strings.TrimSpace(flagURL) != "" {
		cfg.ServerURL = strings.TrimSpace(flagURL)
	}
	if strings.TrimSpace(flagAPIKey) != "" {
		cfg.APIKey = strings.TrimSpace(flagAPIKey)
	}

	cfg.ServerURL = strings.TrimRight(cfg.ServerURL, "/")
	return cfg
}

func loadConfigFile() (*fileConfig, error) {
	paths := []string{
		".scopesentry.yaml",
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		paths = append(paths, filepath.Join(home, ".config", "scopesentry", "config.yaml"))
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			var fc fileConfig
			if err := yaml.Unmarshal(data, &fc); err == nil {
				return &fc, nil
			}
		}
	}
	return nil, os.ErrNotExist
}
