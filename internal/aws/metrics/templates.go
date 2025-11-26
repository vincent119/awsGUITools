package metrics

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// ResourceKind 表示資源類型。
type ResourceKind string

const (
	KindEC2    ResourceKind = "ec2"
	KindRDS    ResourceKind = "rds"
	KindS3     ResourceKind = "s3"
	KindLambda ResourceKind = "lambda"
)

// DefaultQueries 根據資源類型與 ID 回傳預設 KPI 查詢。
func DefaultQueries(kind ResourceKind, resourceID string) []Query {
	switch kind {
	case KindEC2:
		return ec2Queries(resourceID)
	case KindRDS:
		return rdsQueries(resourceID)
	case KindS3:
		return s3Queries(resourceID)
	case KindLambda:
		return lambdaQueries(resourceID)
	default:
		return nil
	}
}

func ec2Queries(instanceID string) []Query {
	dim := []types.Dimension{{Name: aws.String("InstanceId"), Value: aws.String(instanceID)}}
	return []Query{
		{ID: "cpu", MetricName: "CPUUtilization", Namespace: "AWS/EC2", Stat: "Average", Dimensions: dim},
		{ID: "netin", MetricName: "NetworkIn", Namespace: "AWS/EC2", Stat: "Sum", Dimensions: dim},
		{ID: "netout", MetricName: "NetworkOut", Namespace: "AWS/EC2", Stat: "Sum", Dimensions: dim},
		{ID: "diskread", MetricName: "DiskReadBytes", Namespace: "AWS/EC2", Stat: "Sum", Dimensions: dim},
		{ID: "diskwrite", MetricName: "DiskWriteBytes", Namespace: "AWS/EC2", Stat: "Sum", Dimensions: dim},
	}
}

func rdsQueries(dbInstanceID string) []Query {
	dim := []types.Dimension{{Name: aws.String("DBInstanceIdentifier"), Value: aws.String(dbInstanceID)}}
	return []Query{
		{ID: "cpu", MetricName: "CPUUtilization", Namespace: "AWS/RDS", Stat: "Average", Dimensions: dim},
		{ID: "conns", MetricName: "DatabaseConnections", Namespace: "AWS/RDS", Stat: "Sum", Dimensions: dim},
		{ID: "freemem", MetricName: "FreeableMemory", Namespace: "AWS/RDS", Stat: "Average", Dimensions: dim},
		{ID: "readiops", MetricName: "ReadIOPS", Namespace: "AWS/RDS", Stat: "Average", Dimensions: dim},
		{ID: "writeiops", MetricName: "WriteIOPS", Namespace: "AWS/RDS", Stat: "Average", Dimensions: dim},
	}
}

func s3Queries(bucketName string) []Query {
	dim := []types.Dimension{
		{Name: aws.String("BucketName"), Value: aws.String(bucketName)},
		{Name: aws.String("StorageType"), Value: aws.String("StandardStorage")},
	}
	return []Query{
		{ID: "size", MetricName: "BucketSizeBytes", Namespace: "AWS/S3", Stat: "Average", Dimensions: dim},
		{ID: "objects", MetricName: "NumberOfObjects", Namespace: "AWS/S3", Stat: "Average", Dimensions: dim},
	}
}

func lambdaQueries(functionName string) []Query {
	dim := []types.Dimension{{Name: aws.String("FunctionName"), Value: aws.String(functionName)}}
	return []Query{
		{ID: "invocations", MetricName: "Invocations", Namespace: "AWS/Lambda", Stat: "Sum", Dimensions: dim},
		{ID: "errors", MetricName: "Errors", Namespace: "AWS/Lambda", Stat: "Sum", Dimensions: dim},
		{ID: "duration", MetricName: "Duration", Namespace: "AWS/Lambda", Stat: "Average", Dimensions: dim},
		{ID: "throttles", MetricName: "Throttles", Namespace: "AWS/Lambda", Stat: "Sum", Dimensions: dim},
		{ID: "concurrent", MetricName: "ConcurrentExecutions", Namespace: "AWS/Lambda", Stat: "Maximum", Dimensions: dim},
	}
}
