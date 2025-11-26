package widgets

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/service/resource"
)

// StatusBar 顯示目前 profile/region/theme/resource 狀態。
type StatusBar struct {
	view *tview.TextView
}

// NewStatusBar 建立狀態列。
func NewStatusBar() *StatusBar {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	view.SetBorder(false)
	return &StatusBar{view: view}
}

// Primitive 回傳元件。
func (s *StatusBar) Primitive() *tview.TextView {
	return s.view
}

// SetStatus 更新顯示。
func (s *StatusBar) SetStatus(profile, region, theme string, kind resource.Kind, count int, message string) {
	text := fmt.Sprintf("[yellow]Profile:[-] %s  [yellow]Region:[-] %s  [yellow]Theme:[-] %s  [yellow]Kind:[-] %s  [yellow]Count:[-] %d  %s",
		emptyFallback(profile, "default"),
		emptyFallback(region, "us-east-1"),
		emptyFallback(theme, "dark"),
		kind,
		count,
		message,
	)
	s.view.SetText(text)
}

func emptyFallback(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
