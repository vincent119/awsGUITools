package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/aws/metrics"
)

// MetricsTab 顯示 CloudWatch 指標。
type MetricsTab struct {
	text *tview.TextView
}

// NewMetricsTab 建立指標 tab。
func NewMetricsTab() *MetricsTab {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	tv.SetBorder(true).SetTitle("Metrics")
	return &MetricsTab{text: tv}
}

// Primitive 回傳 tview 元件。
func (t *MetricsTab) Primitive() *tview.TextView {
	return t.text
}

// SetLoading 顯示載入中。
func (t *MetricsTab) SetLoading() {
	t.text.SetText("[yellow]載入指標中...[-]")
}

// SetError 顯示錯誤。
func (t *MetricsTab) SetError(err error) {
	t.text.SetText(fmt.Sprintf("[red]錯誤：%s[-]", err.Error()))
}

// SetData 顯示指標資料。
func (t *MetricsTab) SetData(data map[string]metrics.Series, start, end time.Time) {
	if len(data) == 0 {
		t.text.SetText("無指標資料")
		return
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[::b]時間範圍：%s ~ %s[::-]\n\n",
		start.Format("01-02 15:04"), end.Format("01-02 15:04")))

	for id, series := range data {
		b.WriteString(fmt.Sprintf("[yellow]%s[-]\n", id))
		if len(series.Points) == 0 {
			b.WriteString("  （無資料點）\n")
			continue
		}
		// 顯示簡易 sparkline（文字版）
		b.WriteString("  ")
		b.WriteString(sparkline(series.Points))
		b.WriteString("\n")
		// 顯示最新值
		latest := series.Points[len(series.Points)-1]
		b.WriteString(fmt.Sprintf("  最新值：%.2f @ %s\n\n",
			latest.Value, latest.Timestamp.Format("15:04:05")))
	}

	t.text.SetText(b.String())
}

// sparkline 產生簡易文字 sparkline。
func sparkline(points []metrics.Point) string {
	if len(points) == 0 {
		return ""
	}
	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	min, max := points[0].Value, points[0].Value
	for _, p := range points {
		if p.Value < min {
			min = p.Value
		}
		if p.Value > max {
			max = p.Value
		}
	}
	rng := max - min
	if rng == 0 {
		rng = 1
	}
	var sb strings.Builder
	for _, p := range points {
		idx := int((p.Value - min) / rng * float64(len(bars)-1))
		if idx >= len(bars) {
			idx = len(bars) - 1
		}
		sb.WriteRune(bars[idx])
	}
	return sb.String()
}
