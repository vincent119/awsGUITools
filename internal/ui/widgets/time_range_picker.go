package widgets

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TimeRange 代表預設時間區間選項。
type TimeRange struct {
	Label    string
	Duration time.Duration
}

// DefaultRanges 提供常用時間區間。
var DefaultRanges = []TimeRange{
	{Label: "15 分鐘", Duration: 15 * time.Minute},
	{Label: "1 小時", Duration: time.Hour},
	{Label: "6 小時", Duration: 6 * time.Hour},
	{Label: "1 天", Duration: 24 * time.Hour},
	{Label: "7 天", Duration: 7 * 24 * time.Hour},
}

// TimeRangePicker 讓使用者快速選擇時間區間。
type TimeRangePicker struct {
	list     *tview.List
	ranges   []TimeRange
	selected int
	onChange func(start, end time.Time)
}

// NewTimeRangePicker 建立時間區間選擇器。
func NewTimeRangePicker() *TimeRangePicker {
	p := &TimeRangePicker{
		list:     tview.NewList(),
		ranges:   DefaultRanges,
		selected: 0,
	}
	p.list.SetBorder(true).SetTitle("時間區間")
	for i, r := range p.ranges {
		idx := i
		p.list.AddItem(r.Label, "", rune('1'+i), func() {
			p.selected = idx
			p.notifyChange()
		})
	}
	p.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	return p
}

// Primitive 回傳 tview 元件。
func (p *TimeRangePicker) Primitive() *tview.List {
	return p.list
}

// SetOnChange 註冊選擇變更回呼。
func (p *TimeRangePicker) SetOnChange(fn func(start, end time.Time)) {
	p.onChange = fn
}

// Selected 回傳目前選取的時間區間。
func (p *TimeRangePicker) Selected() (start, end time.Time) {
	end = time.Now().UTC()
	start = end.Add(-p.ranges[p.selected].Duration)
	return
}

// SelectedLabel 回傳目前選取的標籤。
func (p *TimeRangePicker) SelectedLabel() string {
	if p.selected >= 0 && p.selected < len(p.ranges) {
		return p.ranges[p.selected].Label
	}
	return ""
}

func (p *TimeRangePicker) notifyChange() {
	if p.onChange != nil {
		start, end := p.Selected()
		p.onChange(start, end)
	}
}

// FormatRange 回傳可讀的時間區間字串。
func FormatRange(start, end time.Time) string {
	return fmt.Sprintf("%s ~ %s", start.Format("01-02 15:04"), end.Format("01-02 15:04"))
}
