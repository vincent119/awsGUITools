package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/vin/ck123gogo/internal/aws/session"
)

// Factory 根據 profile/region 產生各 AWS 服務的 client。
type Factory struct {
	loader session.Loader
}

// NewFactory 建立 Factory 實例。
func NewFactory(loader session.Loader) *Factory {
	return &Factory{loader: loader}
}

func (f *Factory) load(ctx context.Context, profile, region string) (awsCfg aws.Config, err error) {
	if f.loader == nil {
		return aws.Config{}, fmt.Errorf("session loader is nil")
	}
	return f.loader.Config(ctx, profile, region)
}

// EC2 回傳 ec2.Client。
func (f *Factory) EC2(ctx context.Context, profile, region string) (*ec2.Client, error) {
	cfg, err := f.load(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

// RDS 回傳 rds.Client。
func (f *Factory) RDS(ctx context.Context, profile, region string) (*rds.Client, error) {
	cfg, err := f.load(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	return rds.NewFromConfig(cfg), nil
}

// S3 回傳 s3.Client。
func (f *Factory) S3(ctx context.Context, profile, region string) (*s3.Client, error) {
	cfg, err := f.load(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

// Lambda 回傳 lambda.Client。
func (f *Factory) Lambda(ctx context.Context, profile, region string) (*lambda.Client, error) {
	cfg, err := f.load(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	return lambda.NewFromConfig(cfg), nil
}

// CloudWatch 回傳 cloudwatch.Client。
func (f *Factory) CloudWatch(ctx context.Context, profile, region string) (*cloudwatch.Client, error) {
	cfg, err := f.load(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	return cloudwatch.NewFromConfig(cfg), nil
}

// CloudWatchLogs 回傳 cloudwatchlogs.Client。
func (f *Factory) CloudWatchLogs(ctx context.Context, profile, region string) (*cloudwatchlogs.Client, error) {
	cfg, err := f.load(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	return cloudwatchlogs.NewFromConfig(cfg), nil
}
