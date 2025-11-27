// Package tags 提供 AWS 資源標籤管理功能。
package tags

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ResourceKind 表示資源類型。
type ResourceKind string

const (
	KindEC2    ResourceKind = "ec2"
	KindRDS    ResourceKind = "rds"
	KindS3     ResourceKind = "s3"
	KindLambda ResourceKind = "lambda"
)

// EC2TagAPI 定義 EC2 標籤操作介面。
type EC2TagAPI interface {
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, params *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// RDSTagAPI 定義 RDS 標籤操作介面。
type RDSTagAPI interface {
	AddTagsToResource(ctx context.Context, params *rds.AddTagsToResourceInput, optFns ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *rds.RemoveTagsFromResourceInput, optFns ...func(*rds.Options)) (*rds.RemoveTagsFromResourceOutput, error)
}

// S3TagAPI 定義 S3 標籤操作介面。
type S3TagAPI interface {
	PutBucketTagging(ctx context.Context, params *s3.PutBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.PutBucketTaggingOutput, error)
	DeleteBucketTagging(ctx context.Context, params *s3.DeleteBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketTaggingOutput, error)
}

// LambdaTagAPI 定義 Lambda 標籤操作介面。
type LambdaTagAPI interface {
	TagResource(ctx context.Context, params *lambda.TagResourceInput, optFns ...func(*lambda.Options)) (*lambda.TagResourceOutput, error)
	UntagResource(ctx context.Context, params *lambda.UntagResourceInput, optFns ...func(*lambda.Options)) (*lambda.UntagResourceOutput, error)
}

// Repository 封裝多資源標籤操作。
type Repository struct {
	ec2Client    EC2TagAPI
	rdsClient    RDSTagAPI
	s3Client     S3TagAPI
	lambdaClient LambdaTagAPI
}

// NewRepository 建立標籤 Repository。
func NewRepository(ec2Client EC2TagAPI, rdsClient RDSTagAPI, s3Client S3TagAPI, lambdaClient LambdaTagAPI) *Repository {
	return &Repository{
		ec2Client:    ec2Client,
		rdsClient:    rdsClient,
		s3Client:     s3Client,
		lambdaClient: lambdaClient,
	}
}

// CreateTags 新增標籤。
func (r *Repository) CreateTags(ctx context.Context, kind ResourceKind, resourceID string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	// 驗證標籤
	if errs := ValidateTags(tags); len(errs) > 0 {
		return errs[0]
	}

	switch kind {
	case KindEC2:
		return r.createEC2Tags(ctx, resourceID, tags)
	case KindRDS:
		return r.createRDSTags(ctx, resourceID, tags)
	case KindS3:
		return r.createS3Tags(ctx, resourceID, tags)
	case KindLambda:
		return r.createLambdaTags(ctx, resourceID, tags)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// DeleteTags 刪除標籤。
func (r *Repository) DeleteTags(ctx context.Context, kind ResourceKind, resourceID string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	switch kind {
	case KindEC2:
		return r.deleteEC2Tags(ctx, resourceID, keys)
	case KindRDS:
		return r.deleteRDSTags(ctx, resourceID, keys)
	case KindS3:
		return r.deleteS3Tags(ctx, resourceID)
	case KindLambda:
		return r.deleteLambdaTags(ctx, resourceID, keys)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

func (r *Repository) createEC2Tags(ctx context.Context, resourceID string, tags map[string]string) error {
	if r.ec2Client == nil {
		return errors.New("ec2 client is nil")
	}
	var ec2Tags []ec2types.Tag
	for k, v := range tags {
		ec2Tags = append(ec2Tags, ec2types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	_, err := r.ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{resourceID},
		Tags:      ec2Tags,
	})
	return err
}

func (r *Repository) deleteEC2Tags(ctx context.Context, resourceID string, keys []string) error {
	if r.ec2Client == nil {
		return errors.New("ec2 client is nil")
	}
	var ec2Tags []ec2types.Tag
	for _, k := range keys {
		ec2Tags = append(ec2Tags, ec2types.Tag{Key: aws.String(k)})
	}
	_, err := r.ec2Client.DeleteTags(ctx, &ec2.DeleteTagsInput{
		Resources: []string{resourceID},
		Tags:      ec2Tags,
	})
	return err
}

func (r *Repository) createRDSTags(ctx context.Context, resourceARN string, tags map[string]string) error {
	if r.rdsClient == nil {
		return errors.New("rds client is nil")
	}
	var rdsTags []rdstypes.Tag
	for k, v := range tags {
		rdsTags = append(rdsTags, rdstypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	_, err := r.rdsClient.AddTagsToResource(ctx, &rds.AddTagsToResourceInput{
		ResourceName: aws.String(resourceARN),
		Tags:         rdsTags,
	})
	return err
}

func (r *Repository) deleteRDSTags(ctx context.Context, resourceARN string, keys []string) error {
	if r.rdsClient == nil {
		return errors.New("rds client is nil")
	}
	_, err := r.rdsClient.RemoveTagsFromResource(ctx, &rds.RemoveTagsFromResourceInput{
		ResourceName: aws.String(resourceARN),
		TagKeys:      keys,
	})
	return err
}

func (r *Repository) createS3Tags(ctx context.Context, bucketName string, tags map[string]string) error {
	if r.s3Client == nil {
		return errors.New("s3 client is nil")
	}
	var s3Tags []s3types.Tag
	for k, v := range tags {
		s3Tags = append(s3Tags, s3types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	_, err := r.s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucketName),
		Tagging: &s3types.Tagging{
			TagSet: s3Tags,
		},
	})
	return err
}

func (r *Repository) deleteS3Tags(ctx context.Context, bucketName string) error {
	if r.s3Client == nil {
		return errors.New("s3 client is nil")
	}
	_, err := r.s3Client.DeleteBucketTagging(ctx, &s3.DeleteBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

func (r *Repository) createLambdaTags(ctx context.Context, functionARN string, tags map[string]string) error {
	if r.lambdaClient == nil {
		return errors.New("lambda client is nil")
	}
	_, err := r.lambdaClient.TagResource(ctx, &lambda.TagResourceInput{
		Resource: aws.String(functionARN),
		Tags:     tags,
	})
	return err
}

func (r *Repository) deleteLambdaTags(ctx context.Context, functionARN string, keys []string) error {
	if r.lambdaClient == nil {
		return errors.New("lambda client is nil")
	}
	_, err := r.lambdaClient.UntagResource(ctx, &lambda.UntagResourceInput{
		Resource: aws.String(functionARN),
		TagKeys:  keys,
	})
	return err
}
