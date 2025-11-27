// Package logs 提供 CloudWatch Logs 查詢功能。
package logs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// LogsAPI 定義 CloudWatch Logs 呼叫介面，便於測試。
type LogsAPI interface {
	FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

// Fetcher 包裝 CloudWatch Logs 查詢。
type Fetcher struct {
	client LogsAPI
}

// NewFetcher 建立 logs fetcher。
func NewFetcher(client LogsAPI) *Fetcher {
	return &Fetcher{client: client}
}

// Options 控制查詢條件。
type Options struct {
	LogGroup  string
	LogStream string
	Filter    string
	StartTime int64
	EndTime   int64
	Limit     int32
	NextToken *string
}

// Page 結果頁。
type Page struct {
	Events    []types.FilteredLogEvent
	NextToken *string
}

// Filter 查詢 log events。
func (f *Fetcher) Filter(ctx context.Context, opt Options) (Page, error) {
	if f.client == nil {
		return Page{}, fmt.Errorf("cloudwatch logs client is nil")
	}
	if opt.LogGroup == "" {
		return Page{}, fmt.Errorf("log group is required")
	}
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(opt.LogGroup),
		FilterPattern: func() *string {
			if opt.Filter == "" {
				return nil
			}
			return aws.String(opt.Filter)
		}(),
	}
	if opt.Limit > 0 {
		input.Limit = aws.Int32(opt.Limit)
	}
	if opt.LogStream != "" {
		input.LogStreamNames = []string{opt.LogStream}
	}
	if opt.StartTime > 0 {
		input.StartTime = aws.Int64(opt.StartTime)
	}
	if opt.EndTime > 0 {
		input.EndTime = aws.Int64(opt.EndTime)
	}
	if opt.NextToken != nil {
		input.NextToken = opt.NextToken
	}

	resp, err := f.client.FilterLogEvents(ctx, input)
	if err != nil {
		return Page{}, err
	}

	return Page{
		Events:    resp.Events,
		NextToken: resp.NextToken,
	}, nil
}
