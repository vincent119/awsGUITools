package widgets

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/i18n"
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
	// 快捷鍵提示（使用 <key:label> 格式避免被當成顏色標籤）
	shortcuts := fmt.Sprintf("[darkcyan]<?:%s> <q:%s>[-]",
		i18n.T("shortcut.help"),
		i18n.T("shortcut.quit"),
	)
	// 狀態資訊
	status := fmt.Sprintf("[yellow]P:[-]%s [yellow]R:[-]%s [yellow]T:[-]%s [yellow]K:[-]%s [yellow]#:[-]%d",
		emptyFallback(profile, "default"),
		emptyFallback(region, "us-east-1"),
		emptyFallback(theme, "dark"),
		kind,
		count,
	)
	// 組合
	text := shortcuts + " " + status
	// 訊息（放最後，超長時會被截斷）
	if message != "" {
		text += "  " + message
	}
	s.view.SetText(text)
}

func emptyFallback(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
