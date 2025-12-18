package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/vincent119/awsGUITools/internal/tags"
)

// mockEC2TagClient 實作 tags.EC2TagAPI。
type mockEC2TagClient struct {
	createCalled bool
	deleteCalled bool
}

func (m *mockEC2TagClient) CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	m.createCalled = true
	return &ec2.CreateTagsOutput{}, nil
}

func (m *mockEC2TagClient) DeleteTags(ctx context.Context, params *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error) {
	m.deleteCalled = true
	return &ec2.DeleteTagsOutput{}, nil
}

// mockRDSTagClient 實作 tags.RDSTagAPI。
type mockRDSTagClient struct {
	addCalled    bool
	removeCalled bool
}

func (m *mockRDSTagClient) AddTagsToResource(ctx context.Context, params *rds.AddTagsToResourceInput, optFns ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error) {
	m.addCalled = true
	return &rds.AddTagsToResourceOutput{}, nil
}

func (m *mockRDSTagClient) RemoveTagsFromResource(ctx context.Context, params *rds.RemoveTagsFromResourceInput, optFns ...func(*rds.Options)) (*rds.RemoveTagsFromResourceOutput, error) {
	m.removeCalled = true
	return &rds.RemoveTagsFromResourceOutput{}, nil
}

// mockS3TagClient 實作 tags.S3TagAPI。
type mockS3TagClient struct {
	putCalled    bool
	deleteCalled bool
}

func (m *mockS3TagClient) PutBucketTagging(ctx context.Context, params *s3.PutBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
	m.putCalled = true
	return &s3.PutBucketTaggingOutput{}, nil
}

func (m *mockS3TagClient) DeleteBucketTagging(ctx context.Context, params *s3.DeleteBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketTaggingOutput, error) {
	m.deleteCalled = true
	return &s3.DeleteBucketTaggingOutput{}, nil
}

// mockLambdaTagClient 實作 tags.LambdaTagAPI。
type mockLambdaTagClient struct {
	tagCalled   bool
	untagCalled bool
}

func (m *mockLambdaTagClient) TagResource(ctx context.Context, params *lambda.TagResourceInput, optFns ...func(*lambda.Options)) (*lambda.TagResourceOutput, error) {
	m.tagCalled = true
	return &lambda.TagResourceOutput{}, nil
}

func (m *mockLambdaTagClient) UntagResource(ctx context.Context, params *lambda.UntagResourceInput, optFns ...func(*lambda.Options)) (*lambda.UntagResourceOutput, error) {
	m.untagCalled = true
	return &lambda.UntagResourceOutput{}, nil
}

func TestTagsRepository_CreateEC2Tags(t *testing.T) {
	ec2Mock := &mockEC2TagClient{}
	repo := tags.NewRepository(ec2Mock, nil, nil, nil)

	err := repo.CreateTags(context.Background(), tags.KindEC2, "i-12345", map[string]string{"Name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ec2Mock.createCalled {
		t.Error("CreateTags was not called")
	}
}

func TestTagsRepository_DeleteEC2Tags(t *testing.T) {
	ec2Mock := &mockEC2TagClient{}
	repo := tags.NewRepository(ec2Mock, nil, nil, nil)

	err := repo.DeleteTags(context.Background(), tags.KindEC2, "i-12345", []string{"Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ec2Mock.deleteCalled {
		t.Error("DeleteTags was not called")
	}
}

func TestTagsRepository_CreateRDSTags(t *testing.T) {
	rdsMock := &mockRDSTagClient{}
	repo := tags.NewRepository(nil, rdsMock, nil, nil)

	err := repo.CreateTags(context.Background(), tags.KindRDS, "arn:aws:rds:...", map[string]string{"Env": "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rdsMock.addCalled {
		t.Error("AddTagsToResource was not called")
	}
}

func TestTagsRepository_CreateS3Tags(t *testing.T) {
	s3Mock := &mockS3TagClient{}
	repo := tags.NewRepository(nil, nil, s3Mock, nil)

	err := repo.CreateTags(context.Background(), tags.KindS3, "my-bucket", map[string]string{"Project": "demo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s3Mock.putCalled {
		t.Error("PutBucketTagging was not called")
	}
}

func TestTagsRepository_CreateLambdaTags(t *testing.T) {
	lambdaMock := &mockLambdaTagClient{}
	repo := tags.NewRepository(nil, nil, nil, lambdaMock)

	err := repo.CreateTags(context.Background(), tags.KindLambda, "arn:aws:lambda:...", map[string]string{"Team": "dev"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !lambdaMock.tagCalled {
		t.Error("TagResource was not called")
	}
}

func TestTagsRepository_ValidationError(t *testing.T) {
	ec2Mock := &mockEC2TagClient{}
	repo := tags.NewRepository(ec2Mock, nil, nil, nil)

	// 使用 aws: 前綴應該失敗
	err := repo.CreateTags(context.Background(), tags.KindEC2, "i-12345", map[string]string{"aws:reserved": "value"})
	if err == nil {
		t.Error("expected validation error for aws: prefix")
	}
	if ec2Mock.createCalled {
		t.Error("CreateTags should not be called for invalid tags")
	}
}
