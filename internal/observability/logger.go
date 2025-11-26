package observability

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// NewLogger 依需求建立 slog.Logger。
func NewLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

// AWSCallMetrics 用於記錄 AWS SDK 呼叫的延遲與結果，後續可擴充為 OTEL。
type AWSCallMetrics struct {
	logger *slog.Logger
}

// NewAWSCallMetrics 建立度量記錄元件。
func NewAWSCallMetrics(logger *slog.Logger) *AWSCallMetrics {
	return &AWSCallMetrics{logger: logger}
}

// Observe 記錄單次 AWS 呼叫結果。
func (m *AWSCallMetrics) Observe(ctx context.Context, service, operation string, duration time.Duration, err error) {
	if m == nil || m.logger == nil {
		return
	}

	attrs := []any{
		slog.String("service", service),
		slog.String("operation", operation),
		slog.Duration("duration", duration),
	}
	if err != nil {
		m.logger.WarnContext(ctx, "aws call failed", append(attrs, slog.String("error", err.Error()))...)
		return
	}
	m.logger.DebugContext(ctx, "aws call succeed", attrs...)
}
