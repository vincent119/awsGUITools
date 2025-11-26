package list

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/models"
)

// View 負責呈現資源清單。
type View struct {
	table    *tview.Table
	items    []models.ListItem
	onSelect func(models.ListItem)
}

// NewView 建立清單畫面。
func NewView() *View {
	table := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)
	table.SetBorder(true).SetTitle("資源清單")

	v := &View{table: table}
	table.SetSelectedFunc(func(row, _ int) {
		if row <= 0 || row-1 >= len(v.items) {
			return
		}
		if v.onSelect != nil {
			v.onSelect(v.items[row-1])
		}
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := table.GetSelection()
			if row <= 0 || row-1 >= len(v.items) {
				return nil
			}
			if v.onSelect != nil {
				v.onSelect(v.items[row-1])
			}
			return nil
		default:
			return event
		}
	})
	return v
}

// Primitive 回傳 tview 元件。
func (v *View) Primitive() *tview.Table {
	return v.table
}

// SetItems 更新清單內容。
func (v *View) SetItems(items []models.ListItem) {
	v.items = items
	v.table.Clear()

	headers := []string{"名稱", "類型", "狀態", "區域/資訊"}
	for col, header := range headers {
		v.table.SetCell(0, col, headerCell(header))
	}

	for i, item := range items {
		row := i + 1
		v.table.SetCell(row, 0, textCell(item.Name))
		v.table.SetCell(row, 1, textCell(item.Type))
		v.table.SetCell(row, 2, textCell(item.Status))
		region := item.Region
		if region == "" && item.Metadata != nil {
			if ep, ok := item.Metadata["endpoint"]; ok {
				region = ep
			}
		}
		v.table.SetCell(row, 3, textCell(region))
	}

	if len(items) > 0 {
		v.table.Select(1, 0)
	}
}

// SetOnSelect 設定選取事件。
func (v *View) SetOnSelect(fn func(models.ListItem)) {
	v.onSelect = fn
}

// CurrentItem 回傳目前選取項目。
func (v *View) CurrentItem() (models.ListItem, bool) {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.items) {
		return models.ListItem{}, false
	}
	return v.items[row-1], true
}

// Count 回傳列表項目數。
func (v *View) Count() int {
	return len(v.items)
}

func headerCell(text string) *tview.TableCell {
	return tview.NewTableCell(fmt.Sprintf("[::b]%s", text)).
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false)
}

func textCell(text string) *tview.TableCell {
	return tview.NewTableCell(text)
}
