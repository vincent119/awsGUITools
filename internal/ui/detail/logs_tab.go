package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/rivo/tview"
)

// LogsTab 顯示 CloudWatch Logs。
type LogsTab struct {
	text *tview.TextView
}

// NewLogsTab 建立日誌 tab。
func NewLogsTab() *LogsTab {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	tv.SetBorder(true).SetTitle("Logs")
	return &LogsTab{text: tv}
}

// Primitive 回傳 tview 元件。
func (t *LogsTab) Primitive() *tview.TextView {
	return t.text
}

// SetLoading 顯示載入中。
func (t *LogsTab) SetLoading() {
	t.text.SetText("[yellow]載入日誌中...[-]")
}

// SetError 顯示錯誤。
func (t *LogsTab) SetError(err error) {
	t.text.SetText(fmt.Sprintf("[red]錯誤：%s[-]", err.Error()))
}

// SetData 顯示日誌資料。
func (t *LogsTab) SetData(events []types.FilteredLogEvent, hasMore bool) {
	if len(events) == 0 {
		t.text.SetText("無日誌資料")
		return
	}

	var b strings.Builder
	for _, ev := range events {
		ts := time.UnixMilli(aws.ToInt64(ev.Timestamp)).Format("01-02 15:04:05")
		msg := aws.ToString(ev.Message)
		b.WriteString(fmt.Sprintf("[gray]%s[-] %s\n", ts, msg))
	}
	if hasMore {
		b.WriteString("\n[yellow]還有更多日誌...[-]")
	}
	t.text.SetText(b.String())
}
