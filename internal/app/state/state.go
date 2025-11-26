package state

import (
	"sync"
)

// Store 保存應用的全域狀態，例如 profile/region/theme 與搜尋條件。
type Store struct {
	mu      sync.RWMutex
	profile string
	region  string
	theme   string
	filter  string
}

// New 建立狀態儲存。
func New(profile, region, theme string) *Store {
	return &Store{
		profile: profile,
		region:  region,
		theme:   theme,
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

func (s *Store) Filter() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filter
}

func (s *Store) SetProfile(profile string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if profile != "" {
		s.profile = profile
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
