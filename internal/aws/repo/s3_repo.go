package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/vincent119/awsGUITools/internal/models"
)

type S3Repository struct{}

func NewS3Repository() *S3Repository {
	return &S3Repository{}
}

// ListObjects 列出 bucket 中指定 prefix 下的物件與子目錄。
func (r *S3Repository) ListObjects(ctx context.Context, client *s3.Client, bucket, prefix string) ([]models.S3Object, error) {
	if client == nil {
		return nil, fmt.Errorf("s3 client is nil")
	}

	// 確保 prefix 以 / 結尾（除了空字串）
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(1000),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	var objects []models.S3Object

	// 先加入目錄（CommonPrefixes）
	for _, cp := range resp.CommonPrefixes {
		dirName := aws.ToString(cp.Prefix)
		// 移除前綴，只顯示相對路徑
		displayName := strings.TrimPrefix(dirName, prefix)
		displayName = strings.TrimSuffix(displayName, "/")
		if displayName != "" {
			objects = append(objects, models.S3Object{
				Key:         dirName,
				IsDirectory: true,
			})
		}
	}

	// 加入物件
	for _, obj := range resp.Contents {
		key := aws.ToString(obj.Key)
		// 跳過 prefix 本身
		if key == prefix {
			continue
		}
		objects = append(objects, models.S3Object{
			Key:          key,
			Size:         aws.ToInt64(obj.Size),
			LastModified: obj.LastModified.Format("2006-01-02 15:04:05"),
			StorageClass: string(obj.StorageClass),
			IsDirectory:  false,
		})
	}

	return objects, nil
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
