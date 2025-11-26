// Package modals 提供 UI 彈窗元件。
package modals

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ConfirmModal 顯示操作確認對話框。
type ConfirmModal struct {
	modal    *tview.Modal
	onResult func(confirmed bool)
}

// NewConfirmModal 建立確認對話框。
func NewConfirmModal() *ConfirmModal {
	modal := tview.NewModal().
		AddButtons([]string{"確認", "取消"})
	m := &ConfirmModal{modal: modal}
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if m.onResult != nil {
			m.onResult(buttonIndex == 0)
		}
	})
	return m
}

// Primitive 回傳 tview 元件。
func (m *ConfirmModal) Primitive() *tview.Modal {
	return m.modal
}

// Show 顯示確認訊息。
func (m *ConfirmModal) Show(title, message string, onResult func(confirmed bool)) {
	m.modal.SetText(fmt.Sprintf("[::b]%s[::-]\n\n%s", title, message))
	m.onResult = onResult
}

// ResultModal 顯示操作結果（成功/失敗）。
type ResultModal struct {
	modal *tview.Modal
	onOK  func()
}

// NewResultModal 建立結果對話框。
func NewResultModal() *ResultModal {
	modal := tview.NewModal().
		AddButtons([]string{"確定"})
	m := &ResultModal{modal: modal}
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if m.onOK != nil {
			m.onOK()
		}
	})
	return m
}

// Primitive 回傳 tview 元件。
func (m *ResultModal) Primitive() *tview.Modal {
	return m.modal
}

// ShowSuccess 顯示成功訊息。
func (m *ResultModal) ShowSuccess(message string, onOK func()) {
	m.modal.SetText(fmt.Sprintf("[green]✓ 成功[-]\n\n%s", message))
	m.onOK = onOK
}

// ShowError 顯示錯誤訊息。
func (m *ResultModal) ShowError(err error, onOK func()) {
	m.modal.SetText(fmt.Sprintf("[red]✗ 錯誤[-]\n\n%s", err.Error()))
	m.onOK = onOK
}

// ShowInfo 顯示資訊訊息。
func (m *ResultModal) ShowInfo(message string, onOK func()) {
	m.modal.SetText(fmt.Sprintf("[yellow]ℹ 資訊[-]\n\n%s", message))
	m.onOK = onOK
}

// ActionPanel 顯示可執行操作列表。
type ActionPanel struct {
	list     *tview.List
	onAction func(action string)
}

// NewActionPanel 建立操作面板。
func NewActionPanel() *ActionPanel {
	list := tview.NewList().
		ShowSecondaryText(false)
	list.SetBorder(true).SetTitle("操作")
	p := &ActionPanel{list: list}
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if p.onAction != nil {
				p.onAction("")
			}
			return nil
		}
		return event
	})
	return p
}

// Primitive 回傳 tview 元件。
func (p *ActionPanel) Primitive() *tview.List {
	return p.list
}

// SetActions 設定可用操作。
func (p *ActionPanel) SetActions(actions []string, onAction func(action string)) {
	p.list.Clear()
	p.onAction = onAction
	for i, action := range actions {
		act := action
		shortcut := rune('1' + i)
		p.list.AddItem(action, "", shortcut, func() {
			if p.onAction != nil {
				p.onAction(act)
			}
		})
	}
	p.list.AddItem("取消", "", 'q', func() {
		if p.onAction != nil {
			p.onAction("")
		}
	})
}

// AvailableActions 根據資源類型回傳可用操作。
func AvailableActions(resourceType string) []string {
	switch resourceType {
	case "EC2":
		return []string{"啟動", "停止", "重新啟動"}
	case "RDS":
		return []string{"啟動", "停止", "重新啟動"}
	case "Lambda":
		return []string{"測試呼叫"}
	case "S3":
		return []string{} // S3 目前無操作
	default:
		return []string{}
	}
}
