package modals

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/aws/profile"
	"github.com/vin/ck123gogo/internal/i18n"
)

// ProfileInfo is an alias for profile.Info for external use.
type ProfileInfo = profile.Info

// ProfilePicker 顯示 AWS Profile 選擇器。
type ProfilePicker struct {
	list     *tview.List
	flex     *tview.Flex
	onSelect func(info profile.Info)
	onCancel func()
	profiles []profile.Info
	current  string
}

// NewProfilePicker 建立 Profile 選擇器。
func NewProfilePicker() *ProfilePicker {
	list := tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true)

	list.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", i18n.T("profile.select.title"))).
		SetTitleAlign(tview.AlignCenter)

	p := &ProfilePicker{list: list}

	// 處理鍵盤事件
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if p.onCancel != nil {
				p.onCancel()
			}
			return nil
		case tcell.KeyEnter:
			// 由 list 的 selected handler 處理
			return event
		}
		// 支援 j/k 上下移動（vim 風格）
		switch event.Rune() {
		case 'j':
			current := p.list.GetCurrentItem()
			if current < p.list.GetItemCount()-1 {
				p.list.SetCurrentItem(current + 1)
			}
			return nil
		case 'k':
			current := p.list.GetCurrentItem()
			if current > 0 {
				p.list.SetCurrentItem(current - 1)
			}
			return nil
		case 'q':
			if p.onCancel != nil {
				p.onCancel()
			}
			return nil
		}
		return event
	})

	// 建立置中的 flex layout
	p.flex = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, 0, 2, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)

	return p
}

// Primitive 回傳 tview 元件。
func (p *ProfilePicker) Primitive() tview.Primitive {
	return p.flex
}

// SetProfiles 設定可選擇的 profiles。
func (p *ProfilePicker) SetProfiles(profiles []profile.Info, current string) {
	p.list.Clear()
	p.profiles = profiles
	p.current = current

	currentIndex := 0
	for i, info := range profiles {
		idx := i
		inf := info

		// 主要文字：profile 名稱（如果是當前選中的，加上標記）
		mainText := info.Name
		if info.Name == current {
			mainText = fmt.Sprintf("[green]▸ %s %s[-]", info.Name, i18n.T("profile.current"))
			currentIndex = i
		}

		// 次要文字：region 資訊
		secondaryText := "  " + i18n.T("profile.region")
		if info.Region != "" {
			secondaryText += info.Region
		} else {
			secondaryText += i18n.T("profile.region.not_set")
		}

		p.list.AddItem(mainText, secondaryText, 0, func() {
			if p.onSelect != nil {
				p.onSelect(p.profiles[idx])
			}
		})

		// 為了閉包正確引用
		_ = inf
	}

	// 將游標移到當前選中的 profile
	p.list.SetCurrentItem(currentIndex)
}

// SetOnSelect 設定選擇回調。
func (p *ProfilePicker) SetOnSelect(onSelect func(info profile.Info)) {
	p.onSelect = onSelect
}

// SetOnCancel 設定取消回調。
func (p *ProfilePicker) SetOnCancel(onCancel func()) {
	p.onCancel = onCancel
}

// UpdateTitle 更新標題（例如顯示 profile 數量）。
func (p *ProfilePicker) UpdateTitle(count int) {
	p.list.SetTitle(fmt.Sprintf(" %s ", i18n.Tf("profile.select.title_count", count)))
}
