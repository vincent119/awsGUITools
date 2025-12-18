// Package search 提供字串搜尋功能的單元測試。
package search_test

import (
	"testing"

	"github.com/vincent119/awsGUITools/internal/search"
)

func TestMatcherMatch(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		target string
		want   bool
	}{
		{name: "empty query matches everything", query: "", target: "i-123", want: true},
		{name: "prefix match", query: "i-", target: "i-abc123", want: true},
		{name: "exact match case-insensitive", query: "Name", target: "name", want: true},
		{name: "substring match", query: "abc", target: "zzabczz", want: true},
		{name: "fuzzy chunks all present", query: "prod asia", target: "prod-app-asia-01", want: true},
		{name: "fuzzy chunk missing", query: "prod asia", target: "prod-eu", want: false},
		{name: "no match", query: "xyz", target: "name", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := search.NewMatcher(tt.query)
			if got := m.Match(tt.target); got != tt.want {
				t.Fatalf("Match(%q, %q) = %v, want %v", tt.query, tt.target, got, tt.want)
			}
		})
	}
}
