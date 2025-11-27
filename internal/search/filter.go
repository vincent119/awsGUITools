// Package search 提供字串搜尋功能。
package search

import (
	"strings"
)

// Matcher 提供前綴、子字串、簡易模糊比對的工具。
type Matcher struct {
	query   string
	chunks  []string
	enabled bool
}

// NewMatcher 建立新的 Matcher。
func NewMatcher(query string) Matcher {
	q := strings.TrimSpace(query)
	return Matcher{
		query:   strings.ToLower(q),
		enabled: q != "",
		chunks:  strings.Fields(strings.ToLower(q)),
	}
}

// Match 回傳 target 是否符合搜尋條件。
func (m Matcher) Match(target string) bool {
	if !m.enabled {
		return true
	}
	targetLower := strings.ToLower(target)

	// 完整匹配或前綴
	if strings.HasPrefix(targetLower, m.query) || targetLower == m.query {
		return true
	}
	// 子字串
	if strings.Contains(targetLower, m.query) {
		return true
	}
	// 簡易模糊：所有分詞都需出現
	for _, chunk := range m.chunks {
		if !strings.Contains(targetLower, chunk) {
			return false
		}
	}
	return true
}
