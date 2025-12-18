package modals

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/vincent119/awsGUITools/internal/tags"
)

// TagsEditor 提供標籤編輯介面。
type TagsEditor struct {
	form     *tview.Form
	list     *tview.List
	flex     *tview.Flex
	tags     map[string]string
	onSave   func(added, removed map[string]string)
	onCancel func()
}

// NewTagsEditor 建立標籤編輯器。
func NewTagsEditor() *TagsEditor {
	e := &TagsEditor{
		form: tview.NewForm(),
		list: tview.NewList().ShowSecondaryText(true),
		tags: make(map[string]string),
	}

	e.form.SetBorder(true).SetTitle("新增標籤")
	e.list.SetBorder(true).SetTitle("現有標籤 [d:刪除]")

	e.flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.list, 0, 1, true).
		AddItem(e.form, 7, 0, false)

	e.setupForm()
	e.setupListInput()

	return e
}

// Primitive 回傳 tview 元件。
func (e *TagsEditor) Primitive() *tview.Flex {
	return e.flex
}

// SetTags 設定現有標籤。
func (e *TagsEditor) SetTags(existing map[string]string) {
	e.tags = make(map[string]string, len(existing))
	for k, v := range existing {
		e.tags[k] = v
	}
	e.refreshList()
}

// SetOnSave 註冊儲存回呼。
func (e *TagsEditor) SetOnSave(fn func(added, removed map[string]string)) {
	e.onSave = fn
}

// SetOnCancel 註冊取消回呼。
func (e *TagsEditor) SetOnCancel(fn func()) {
	e.onCancel = fn
}

func (e *TagsEditor) setupForm() {
	var keyInput, valueInput *tview.InputField

	keyInput = tview.NewInputField().
		SetLabel("Key: ").
		SetFieldWidth(30)

	valueInput = tview.NewInputField().
		SetLabel("Value: ").
		SetFieldWidth(30)

	e.form.AddFormItem(keyInput)
	e.form.AddFormItem(valueInput)
	e.form.AddButton("新增", func() {
		key := keyInput.GetText()
		value := valueInput.GetText()

		if err := tags.ValidateTag(key, value); err != nil {
			// 顯示錯誤（簡化處理）
			return
		}

		e.tags[key] = value
		e.refreshList()
		keyInput.SetText("")
		valueInput.SetText("")
	})
	e.form.AddButton("儲存", func() {
		if e.onSave != nil {
			e.onSave(e.tags, nil)
		}
	})
	e.form.AddButton("取消", func() {
		if e.onCancel != nil {
			e.onCancel()
		}
	})
}

func (e *TagsEditor) setupListInput() {
	e.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'd' || event.Key() == tcell.KeyDelete {
			idx := e.list.GetCurrentItem()
			if idx >= 0 && idx < e.list.GetItemCount() {
				key, _ := e.list.GetItemText(idx)
				delete(e.tags, key)
				e.refreshList()
			}
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			if e.onCancel != nil {
				e.onCancel()
			}
			return nil
		}
		return event
	})
}

func (e *TagsEditor) refreshList() {
	e.list.Clear()
	keys := make([]string, 0, len(e.tags))
	for k := range e.tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := e.tags[k]
		e.list.AddItem(k, fmt.Sprintf("  %s", v), 0, nil)
	}
}
