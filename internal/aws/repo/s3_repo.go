package repo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/vin/ck123gogo/internal/models"
)

type S3Repository struct{}

func NewS3Repository() *S3Repository {
	return &S3Repository{}
}

func (r *S3Repository) ListBuckets(ctx context.Context, client *s3.Client) ([]models.S3Bucket, error) {
	if client == nil {
		return nil, fmt.Errorf("s3 client is nil")
	}

	resp, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	var buckets []models.S3Bucket
	for _, bucket := range resp.Buckets {
		name := deref(bucket.Name)
		location, _ := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		versioning, _ := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
			Bucket: bucket.Name,
		})
		encryption, _ := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		})
		lifecycle, _ := client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
			Bucket: bucket.Name,
		})
		policy, _ := client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
			Bucket: bucket.Name,
		})

		buckets = append(buckets, models.S3Bucket{
			Name:       name,
			Region:     resolveLocation(location),
			Versioning: versioningStatus(versioning),
			Encryption: encryptionAlgo(encryption),
			Lifecycle:  lifecycleString(lifecycle),
			Policy:     policyString(policy),
		})
	}

	return buckets, nil
}

func resolveLocation(output *s3.GetBucketLocationOutput) string {
	if output == nil || output.LocationConstraint == "" {
		return "us-east-1"
	}
	return string(output.LocationConstraint)
}

func versioningStatus(output *s3.GetBucketVersioningOutput) string {
	if output == nil {
		return "Disabled"
	}
	return string(output.Status)
}

func encryptionAlgo(output *s3.GetBucketEncryptionOutput) string {
	if output == nil || output.ServerSideEncryptionConfiguration == nil {
		return ""
	}
	rules := output.ServerSideEncryptionConfiguration.Rules
	if len(rules) == 0 || rules[0].ApplyServerSideEncryptionByDefault == nil {
		return ""
	}
	return string(rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm)
}

func lifecycleString(output *s3.GetBucketLifecycleConfigurationOutput) string {
	if output == nil {
		return ""
	}
	return fmt.Sprintf("%d rules", len(output.Rules))
}

func policyString(output *s3.GetBucketPolicyOutput) string {
	if output == nil || output.Policy == nil {
		return ""
	}
	return aws.ToString(output.Policy)
}
