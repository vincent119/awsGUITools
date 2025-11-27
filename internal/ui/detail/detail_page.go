package detail

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/i18n"
	"github.com/vin/ck123gogo/internal/models"
	"github.com/vin/ck123gogo/internal/ui/modals"
)

// View 負責呈現資源詳情。
type View struct {
	text        *tview.TextView
	currentItem *models.ListItem
	onAction    func(resourceType, resourceID, action string)
}

// NewView 建立詳情畫面。
func NewView() *View {
	text := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	text.SetBorder(true).SetTitle(i18n.T("ui.resource_detail"))
	v := &View{text: text}
	text.SetInputCapture(v.handleInput)
	return v
}

// Primitive 回傳 tview 元件。
func (v *View) Primitive() *tview.TextView {
	return v.text
}

// SetCurrentItem 設定目前選取的資源項目。
func (v *View) SetCurrentItem(item *models.ListItem) {
	v.currentItem = item
}

// SetOnAction 註冊操作回呼。
func (v *View) SetOnAction(fn func(resourceType, resourceID, action string)) {
	v.onAction = fn
}

// AvailableActions 回傳目前資源可用的操作。
func (v *View) AvailableActions() []string {
	if v.currentItem == nil {
		return nil
	}
	return modals.AvailableActions(v.currentItem.Type)
}

func (v *View) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// 'a' 鍵觸發操作（由外層 UI 處理彈窗）
	return event
}

// SetDetail 顯示詳細資訊。
func (v *View) SetDetail(detail models.DetailView) {
	if len(detail.Overview) == 0 {
		v.text.SetText(i18n.T("ui.no_resource"))
		return
	}

	var b strings.Builder
	b.WriteString("[::b]Overview[::-]\n")
	for k, val := range detail.Overview {
		if val == "" {
			continue
		}
		b.WriteString(" - ")
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(val)
		b.WriteString("\n")
	}

	if len(detail.Relations) > 0 {
		b.WriteString("\n[::b]Relations[::-]\n")
		for k, list := range detail.Relations {
			if len(list) == 0 {
				continue
			}
			b.WriteString(" - ")
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(strings.Join(list, ", "))
			b.WriteString("\n")
		}
	}

	if len(detail.Tags) > 0 {
		b.WriteString("\n[::b]Tags[::-]\n")
		for k, val := range detail.Tags {
			b.WriteString(" - ")
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(val)
			b.WriteString("\n")
		}
	}

	v.text.SetText(b.String())
}
