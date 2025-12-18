// Package state 提供應用程式的全域狀態管理。
package state

import (
	"sync"

	"github.com/vincent119/awsGUITools/internal/aws/profile"
)

// Store 保存應用的全域狀態，例如 profile/region/theme/language 與搜尋條件。
type Store struct {
	mu       sync.RWMutex
	profile  string
	region   string
	theme    string
	language string
	filter   string
	profiles *profile.List // 可用的 AWS profiles 列表
}

// New 建立狀態儲存。
func New(profileName, region, theme, language string) *Store {
	return &Store{
		profile:  profileName,
		region:   region,
		theme:    theme,
		language: language,
	}
}

// NewWithProfiles 建立狀態儲存，並注入 profiles 列表。
func NewWithProfiles(profileName, region, theme, language string, profiles *profile.List) *Store {
	return &Store{
		profile:  profileName,
		region:   region,
		theme:    theme,
		language: language,
		profiles: profiles,
	}
}

func (s *Store) Profile() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile
}

func (s *Store) Region() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.region
}

func (s *Store) Theme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.theme
}

func (s *Store) Language() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.language
}

func (s *Store) Filter() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filter
}

// SetProfile 設定當前 profile，並自動切換到該 profile 對應的 region（若有）。
func (s *Store) SetProfile(profileName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if profileName == "" {
		return
	}
	s.profile = profileName

	// 自動切換到該 profile 對應的 region
	if s.profiles != nil {
		if info, found := s.profiles.GetProfile(profileName); found && info.Region != "" {
			s.region = info.Region
		}
	}
}

// SetProfileWithoutRegionSwitch 設定 profile 但不自動切換 region。
func (s *Store) SetProfileWithoutRegionSwitch(profileName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if profileName != "" {
		s.profile = profileName
	}
}

func (s *Store) SetRegion(region string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if region != "" {
		s.region = region
	}
}

func (s *Store) SetTheme(theme string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if theme != "" {
		s.theme = theme
	}
}

func (s *Store) SetLanguage(language string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if language != "" {
		s.language = language
	}
}

func (s *Store) SetFilter(filter string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filter = filter
}

// Snapshot 回傳目前 state 資訊，避免 UI 需要多次鎖定。
func (s *Store) Snapshot() (profile, region, theme, filter string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile, s.region, s.theme, s.filter
}

// Profiles 回傳可用的 AWS profiles 列表。
func (s *Store) Profiles() *profile.List {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profiles
}

// SetProfiles 設定可用的 AWS profiles 列表。
func (s *Store) SetProfiles(profiles *profile.List) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles = profiles
}

// ProfileNames 回傳所有可用的 profile 名稱。
func (s *Store) ProfileNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.profiles == nil {
		return nil
	}
	return s.profiles.Names()
}

// GetProfileInfo 根據名稱取得 profile 資訊。
func (s *Store) GetProfileInfo(name string) (profile.Info, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.profiles == nil {
		return profile.Info{}, false
	}
	return s.profiles.GetProfile(name)
}
