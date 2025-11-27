// Package profile 提供 AWS CLI 設定檔解析功能，從 ~/.aws/config 與 ~/.aws/credentials 讀取 profiles。
// 支援跨平台（Windows/macOS/Linux）路徑處理。
package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// Info 描述單一 AWS profile 的資訊。
type Info struct {
	Name   string // profile 名稱
	Region string // 對應的預設 region（可能為空）
}

// List 描述可用的 AWS profiles 列表。
type List struct {
	Profiles []Info
	Default  string // 預設 profile 名稱
}

// Parser 負責解析 AWS CLI 設定檔。
type Parser struct {
	configPath      string
	credentialsPath string
}

// NewParser 建立 Parser，使用預設路徑。
// 跨平台支援：
//   - Linux/macOS: ~/.aws/config, ~/.aws/credentials
//   - Windows: %USERPROFILE%\.aws\config, %USERPROFILE%\.aws\credentials
//
// 環境變數可覆蓋預設路徑：
//   - AWS_CONFIG_FILE: 覆蓋 config 文件路徑
//   - AWS_SHARED_CREDENTIALS_FILE: 覆蓋 credentials 文件路徑
func NewParser() (*Parser, error) {
	configPath, credentialsPath, err := resolveAWSConfigPaths()
	if err != nil {
		return nil, err
	}

	return &Parser{
		configPath:      configPath,
		credentialsPath: credentialsPath,
	}, nil
}

// resolveAWSConfigPaths 根據作業系統和環境變數解析 AWS 設定檔路徑。
func resolveAWSConfigPaths() (configPath, credentialsPath string, err error) {
	// 優先使用環境變數
	configPath = os.Getenv("AWS_CONFIG_FILE")
	credentialsPath = os.Getenv("AWS_SHARED_CREDENTIALS_FILE")

	// 如果沒有設定環境變數，使用預設路徑
	if configPath == "" || credentialsPath == "" {
		homeDir, err := getHomeDir()
		if err != nil {
			return "", "", err
		}

		awsDir := filepath.Join(homeDir, ".aws")

		if configPath == "" {
			configPath = filepath.Join(awsDir, "config")
		}
		if credentialsPath == "" {
			credentialsPath = filepath.Join(awsDir, "credentials")
		}
	}

	return configPath, credentialsPath, nil
}

// getHomeDir 取得使用者 home 目錄，支援跨平台。
func getHomeDir() (string, error) {
	// 優先使用 os.UserHomeDir()，它已經處理跨平台
	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		return homeDir, nil
	}

	// Fallback: 根據作業系統使用不同的環境變數
	switch runtime.GOOS {
	case "windows":
		// Windows 優先使用 USERPROFILE，其次是 HOMEDRIVE + HOMEPATH
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home, nil
		}
		if drive := os.Getenv("HOMEDRIVE"); drive != "" {
			if path := os.Getenv("HOMEPATH"); path != "" {
				return drive + path, nil
			}
		}
	default:
		// Unix-like 系統使用 HOME
		if home := os.Getenv("HOME"); home != "" {
			return home, nil
		}
	}

	return "", fmt.Errorf("cannot determine home directory on %s", runtime.GOOS)
}

// GetConfigPath 回傳目前使用的 config 文件路徑。
func (p *Parser) GetConfigPath() string {
	return p.configPath
}

// GetCredentialsPath 回傳目前使用的 credentials 文件路徑。
func (p *Parser) GetCredentialsPath() string {
	return p.credentialsPath
}

// NewParserWithPaths 建立 Parser，使用指定的設定檔路徑（主要用於測試）。
func NewParserWithPaths(configPath, credentialsPath string) *Parser {
	return &Parser{
		configPath:      configPath,
		credentialsPath: credentialsPath,
	}
}

// Parse 解析 AWS 設定檔，回傳可用的 profiles 列表。
func (p *Parser) Parse() (*List, error) {
	// 從 config 讀取 profiles 與 regions
	configProfiles, err := p.parseConfigFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// 從 credentials 讀取額外的 profiles
	credProfiles, err := p.parseCredentialsFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("parse credentials file: %w", err)
	}

	// 合併 profiles（config 優先，因為包含 region 資訊）
	merged := p.mergeProfiles(configProfiles, credProfiles)

	return merged, nil
}

// parseConfigFile 解析 ~/.aws/config 文件。
// config 文件中 section 格式：
// - [default] -> profile name: "default"
// - [profile xxx] -> profile name: "xxx"
func (p *Parser) parseConfigFile() (map[string]Info, error) {
	file, err := os.Open(p.configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	profiles := make(map[string]Info)
	var currentProfile string

	// 正則匹配 section header
	// [default] 或 [profile xxx]
	sectionRe := regexp.MustCompile(`^\s*\[\s*(profile\s+)?([^\]]+)\s*\]\s*$`)
	// key = value
	kvRe := regexp.MustCompile(`^\s*([^=]+?)\s*=\s*(.+?)\s*$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳過空行和註解
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}

		// 檢查是否為 section header
		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			// Trim whitespace from profile name
			currentProfile = strings.TrimSpace(matches[2])
			if _, exists := profiles[currentProfile]; !exists {
				profiles[currentProfile] = Info{Name: currentProfile}
			}
			continue
		}

		// 解析 key = value
		if currentProfile != "" {
			if matches := kvRe.FindStringSubmatch(line); matches != nil {
				key := strings.ToLower(matches[1])
				value := matches[2]

				if key == "region" {
					info := profiles[currentProfile]
					info.Region = value
					profiles[currentProfile] = info
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan config file: %w", err)
	}

	return profiles, nil
}

// parseCredentialsFile 解析 ~/.aws/credentials 文件。
// credentials 文件中 section 格式直接為 [profile_name]。
func (p *Parser) parseCredentialsFile() (map[string]Info, error) {
	file, err := os.Open(p.credentialsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	profiles := make(map[string]Info)

	// 正則匹配 section header [xxx]
	sectionRe := regexp.MustCompile(`^\s*\[\s*([^\]]+)\s*\]\s*$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳過空行和註解
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}

		// 檢查是否為 section header
		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			// Trim whitespace from profile name
			profileName := strings.TrimSpace(matches[1])
			profiles[profileName] = Info{Name: profileName}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan credentials file: %w", err)
	}

	return profiles, nil
}

// mergeProfiles 合併 config 與 credentials 的 profiles。
// config 中的資訊優先（包含 region），credentials 僅補充 profile 名稱。
func (p *Parser) mergeProfiles(configProfiles, credProfiles map[string]Info) *List {
	merged := make(map[string]Info)

	// 先加入 config profiles（包含 region 資訊）
	for name, info := range configProfiles {
		merged[name] = info
	}

	// 補充 credentials 中獨有的 profiles
	for name, info := range credProfiles {
		if _, exists := merged[name]; !exists {
			merged[name] = info
		}
	}

	// 轉換為 slice 並排序
	list := &List{
		Profiles: make([]Info, 0, len(merged)),
		Default:  "default",
	}

	for _, info := range merged {
		list.Profiles = append(list.Profiles, info)
	}

	// 按名稱排序，但 default 放第一個
	sort.Slice(list.Profiles, func(i, j int) bool {
		if list.Profiles[i].Name == "default" {
			return true
		}
		if list.Profiles[j].Name == "default" {
			return false
		}
		return list.Profiles[i].Name < list.Profiles[j].Name
	})

	return list
}

// GetProfile 根據名稱查找 profile。
func (l *List) GetProfile(name string) (Info, bool) {
	for _, p := range l.Profiles {
		if p.Name == name {
			return p, true
		}
	}
	return Info{}, false
}

// Names 回傳所有 profile 名稱。
func (l *List) Names() []string {
	names := make([]string, len(l.Profiles))
	for i, p := range l.Profiles {
		names[i] = p.Name
	}
	return names
}

// HasProfiles 檢查是否有可用的 profiles。
func (l *List) HasProfiles() bool {
	return len(l.Profiles) > 0
}
