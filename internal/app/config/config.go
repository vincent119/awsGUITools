// Package config 提供 AWS TUI 應用的設定載入與預設值管理。
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 描述應用所需的核心設定。
type Config struct {
	Profile        string        `yaml:"profile"`
	Region         string        `yaml:"region"`
	Theme          string        `yaml:"theme"`
	PageSize       int           `yaml:"page_size"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
}

// Default 產生預設設定。
func Default() Config {
	return Config{
		Profile:        lookupDefault(os.Getenv("AWS_PROFILE"), "default"),
		Region:         lookupDefault(os.Getenv("AWS_REGION"), "us-east-1"),
		Theme:          lookupDefault(os.Getenv("AWS_TUI_THEME"), "dark"),
		PageSize:       parseIntWithDefault(os.Getenv("AWS_TUI_PAGE_SIZE"), 50),
		RequestTimeout: parseDurationWithDefault(os.Getenv("AWS_TUI_TIMEOUT"), 5*time.Second),
	}
}

// Load 根據指定檔案或環境變數載入設定。
func Load(path string) (Config, error) {
	cfg := Default()

	source := path
	if source == "" {
		source = os.Getenv("AWS_TUI_CONFIG")
	}
	if source == "" {
		source = defaultConfigPath()
	}

	if source != "" {
		if err := applyFileOverride(source, &cfg); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

func applyFileOverride(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open config %s: %w", path, err)
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}

	var fileCfg fileConfig
	if err := yaml.Unmarshal(raw, &fileCfg); err != nil {
		return fmt.Errorf("decode config %s: %w", path, err)
	}

	if fileCfg.Profile != "" {
		cfg.Profile = fileCfg.Profile
	}
	if fileCfg.Region != "" {
		cfg.Region = fileCfg.Region
	}
	if fileCfg.Theme != "" {
		cfg.Theme = fileCfg.Theme
	}
	if fileCfg.PageSize != 0 {
		cfg.PageSize = fileCfg.PageSize
	}
	if fileCfg.RequestTimeout != "" {
		cfg.RequestTimeout = parseDurationWithDefault(fileCfg.RequestTimeout, cfg.RequestTimeout)
	}

	return nil
}

type fileConfig struct {
	Profile        string `yaml:"profile"`
	Region         string `yaml:"region"`
	Theme          string `yaml:"theme"`
	PageSize       int    `yaml:"page_size"`
	RequestTimeout string `yaml:"request_timeout"`
}

func defaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "aws-tui", "config.yaml")
}

func lookupDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func parseIntWithDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	v, err := strconv.Atoi(value)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func parseDurationWithDefault(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}
