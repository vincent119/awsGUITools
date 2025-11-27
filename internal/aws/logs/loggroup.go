// Package logs 提供 CloudWatch Logs 相關功能。
package logs

import "fmt"

// ResourceKind 表示資源類型。
type ResourceKind string

const (
	KindEC2    ResourceKind = "ec2"
	KindRDS    ResourceKind = "rds"
	KindS3     ResourceKind = "s3"
	KindLambda ResourceKind = "lambda"
)

// DeriveLogGroup 根據資源類型與 ID 推導 CloudWatch Logs log group 名稱。
// 若無法推導則回傳空字串。
func DeriveLogGroup(kind ResourceKind, resourceID string) string {
	switch kind {
	case KindLambda:
		// Lambda 預設 log group 格式
		return fmt.Sprintf("/aws/lambda/%s", resourceID)
	case KindRDS:
		// RDS 常見 log group（需使用者確認 engine）
		// 這裡提供 MySQL/PostgreSQL 通用格式
		return fmt.Sprintf("/aws/rds/instance/%s/error", resourceID)
	case KindEC2:
		// EC2 需要 CloudWatch Agent，無法自動推導
		return ""
	case KindS3:
		// S3 Server Access Logs 通常放在使用者指定的 bucket，無法自動推導
		return ""
	default:
		return ""
	}
}
