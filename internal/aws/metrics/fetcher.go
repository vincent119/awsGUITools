package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// Query 定義單一 CloudWatch 指標查詢。
type Query struct {
	ID         string
	MetricName string
	Namespace  string
	Stat       string
	Dimensions []types.Dimension
}

// Options 控制查詢時間範圍與取樣。
type Options struct {
	StartTime time.Time
	EndTime   time.Time
	Period    int32
	Queries   []Query
}

// Point 表示單一時間序列點。
type Point struct {
	Timestamp time.Time
	Value     float64
}

// Series 代表同一 Metric 的結果。
type Series struct {
	ID     string
	Points []Point
}

// MetricAPI 抽象化 CloudWatch client，便於測試。
type MetricAPI interface {
	GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error)
}

// Fetcher 封裝 CloudWatch 指標查詢。
type Fetcher struct {
	client MetricAPI
}

// NewFetcher 建立 Fetcher。
func NewFetcher(client MetricAPI) *Fetcher {
	return &Fetcher{client: client}
}

// Fetch 送出指標查詢並回傳每個 Query ID 的結果。
func (f *Fetcher) Fetch(ctx context.Context, opt Options) (map[string]Series, error) {
	if f.client == nil {
		return nil, fmt.Errorf("cloudwatch client is nil")
	}
	if len(opt.Queries) == 0 {
		return nil, fmt.Errorf("no metric queries provided")
	}
	input := &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(opt.StartTime),
		EndTime:   aws.Time(opt.EndTime),
	}
	if opt.Period > 0 {
		input.ScanBy = types.ScanByTimestampAscending
	}

	for _, q := range opt.Queries {
		metricStat := &types.MetricStat{
			Metric: &types.Metric{
				MetricName: aws.String(q.MetricName),
				Namespace:  aws.String(q.Namespace),
				Dimensions: q.Dimensions,
			},
			Stat: aws.String(q.Stat),
		}
		if opt.Period > 0 {
			metricStat.Period = aws.Int32(opt.Period)
		}
		input.MetricDataQueries = append(input.MetricDataQueries, types.MetricDataQuery{
			Id:         aws.String(q.ID),
			MetricStat: metricStat,
		})
	}

	resp, err := f.client.GetMetricData(ctx, input)
	if err != nil {
		return nil, err
	}

	result := make(map[string]Series, len(resp.MetricDataResults))
	for _, data := range resp.MetricDataResults {
		id := aws.ToString(data.Id)
		series := Series{ID: id}
		for i := range data.Timestamps {
			if i >= len(data.Values) {
				break
			}
			series.Points = append(series.Points, Point{
				Timestamp: data.Timestamps[i],
				Value:     data.Values[i],
			})
		}
		result[id] = series
	}
	return result, nil
}
