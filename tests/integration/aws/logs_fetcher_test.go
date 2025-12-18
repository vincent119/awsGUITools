package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	"github.com/vincent119/awsGUITools/internal/aws/logs"
)

type stubLogsClient struct {
	cloudwatchlogs.Client
	Output *cloudwatchlogs.FilterLogEventsOutput
}

func (s *stubLogsClient) FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
	return s.Output, nil
}

func TestLogsFetcher(t *testing.T) {
	stub := &stubLogsClient{
		Output: &cloudwatchlogs.FilterLogEventsOutput{
			Events: []types.FilteredLogEvent{
				{
					Message:       aws.String("test event"),
					LogStreamName: aws.String("stream-1"),
				},
			},
			NextToken: aws.String("next"),
		},
	}

	fetcher := logs.NewFetcher(stub)
	page, err := fetcher.Filter(context.Background(), logs.Options{
		LogGroup: "/aws/lambda/test",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	if len(page.Events) != 1 || aws.ToString(page.NextToken) != "next" {
		t.Fatalf("unexpected page: %#v", page)
	}
}
