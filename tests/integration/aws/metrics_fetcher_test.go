package aws_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/vin/ck123gogo/internal/aws/metrics"
)

type stubCWClient struct {
	cloudwatch.Client
	Output *cloudwatch.GetMetricDataOutput
}

func (s *stubCWClient) GetMetricData(ctx context.Context, params *cloudwatch.GetMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricDataOutput, error) {
	return s.Output, nil
}

func TestFetcherFetch(t *testing.T) {
	now := time.Now()
	stub := &stubCWClient{
		Output: &cloudwatch.GetMetricDataOutput{
			MetricDataResults: []types.MetricDataResult{
				{
					Id: aws.String("m1"),
					Timestamps: []time.Time{
						now.Add(-5 * time.Minute),
						now,
					},
					Values: []float64{10, 20},
				},
			},
		},
	}

	fetcher := metrics.NewFetcher(stub)
	opt := metrics.Options{
		StartTime: now.Add(-10 * time.Minute),
		EndTime:   now,
		Period:    60,
		Queries: []metrics.Query{
			{
				ID:         "m1",
				MetricName: "CPUUtilization",
				Namespace:  "AWS/EC2",
				Stat:       "Average",
			},
		},
	}

	result, err := fetcher.Fetch(context.Background(), opt)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	series, ok := result["m1"]
	if !ok || len(series.Points) != 2 {
		t.Fatalf("unexpected series: %#v", series)
	}
}
