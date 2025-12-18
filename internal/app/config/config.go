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

	"github.com/vincent119/awsGUITools/internal/aws/profile"
	"gopkg.in/yaml.v3"
)

// Config 描述應用所需的核心設定。
type Config struct {
	Profile        string        `yaml:"profile"`
	Region         string        `yaml:"region"`
	Theme          string        `yaml:"theme"`
	Language       string        `yaml:"language"`
	PageSize       int           `yaml:"page_size"`
	RequestTimeout time.Duration `yaml:"request_timeout"`

	// Profiles 儲存從 ~/.aws/config 解析出的 profile 列表
	Profiles *profile.List `yaml:"-"`
}

// Default 產生預設設定，並從 ~/.aws/config 讀取 profiles。
func Default() Config {
	cfg := Config{
		Theme:          lookupDefault(os.Getenv("AWS_TUI_THEME"), "dark"),
		Language:       lookupDefault(os.Getenv("AWS_TUI_LANGUAGE"), "en"), // 預設英文
		PageSize:       parseIntWithDefault(os.Getenv("AWS_TUI_PAGE_SIZE"), 50),
		RequestTimeout: parseDurationWithDefault(os.Getenv("AWS_TUI_TIMEOUT"), 5*time.Second),
	}

	// 從 AWS config 讀取 profiles
	parser, err := profile.NewParser()
	if err == nil {
		profiles, err := parser.Parse()
		if err == nil && profiles.HasProfiles() {
			cfg.Profiles = profiles
			// 設定預設 profile（優先使用環境變數，其次使用 default）
			defaultProfile := lookupDefault(os.Getenv("AWS_PROFILE"), "default")
			if info, found := profiles.GetProfile(defaultProfile); found {
				cfg.Profile = info.Name
				cfg.Region = info.Region
			} else if len(profiles.Profiles) > 0 {
				// 使用第一個可用的 profile
				cfg.Profile = profiles.Profiles[0].Name
				cfg.Region = profiles.Profiles[0].Region
			}
		}
	}

	// 如果 region 仍為空，使用環境變數或預設值
	if cfg.Region == "" {
		cfg.Region = lookupDefault(os.Getenv("AWS_REGION"), "us-east-1")
	}

	// 如果 profile 仍為空，使用環境變數或預設值
	if cfg.Profile == "" {
		cfg.Profile = lookupDefault(os.Getenv("AWS_PROFILE"), "default")
	}

	return cfg
}

// Load 根據指定檔案或環境變數載入設定。
// Profile 與 Region 優先從 ~/.aws/config 讀取，config.yaml 可覆寫非 AWS 相關設定。
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

	// 如果 config.yaml 指定了 profile，嘗試從 profiles 列表中取得對應的 region
	if cfg.Profiles != nil {
		if info, found := cfg.Profiles.GetProfile(cfg.Profile); found && info.Region != "" {
			cfg.Region = info.Region
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

	// Note: profile 和 region 已從 ~/.aws/config 讀取，
	// config.yaml 中的 profile 僅作為選擇哪個 profile 的提示。
	// 實際 region 由該 profile 在 ~/.aws/config 中的設定決定。
	if fileCfg.Profile != "" {
		cfg.Profile = fileCfg.Profile
	}
	// config.yaml 中的 region 作為 fallback（若 profile 沒有設定 region）
	if fileCfg.Region != "" && cfg.Region == "" {
		cfg.Region = fileCfg.Region
	}
	if fileCfg.Theme != "" {
		cfg.Theme = fileCfg.Theme
	}
	if fileCfg.Language != "" {
		cfg.Language = fileCfg.Language
	}
	if fileCfg.PageSize != 0 {
		cfg.PageSize = fileCfg.PageSize
	}
	if fileCfg.RequestTimeout != "" {
		cfg.RequestTimeout = parseDurationWithDefault(fileCfg.RequestTimeout, cfg.RequestTimeout)
	}

	return nil
}

// fileConfig 定義 config.yaml 的結構。
// 注意：profile/region 優先從 ~/.aws/config 讀取。
type fileConfig struct {
	Profile        string `yaml:"profile"`         // 選擇哪個 AWS profile（對應 ~/.aws/config 中的 profile）
	Region         string `yaml:"region"`          // Fallback region（若 profile 沒有設定 region）
	Theme          string `yaml:"theme"`           // UI 主題
	Language       string `yaml:"language"`        // 介面語言：en, zh-TW
	PageSize       int    `yaml:"page_size"`       // 分頁大小
	RequestTimeout string `yaml:"request_timeout"` // 請求超時
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
