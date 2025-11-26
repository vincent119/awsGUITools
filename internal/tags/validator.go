// Package tags 提供 AWS 資源標籤管理功能。
package tags

import (
	"errors"
	"regexp"
	"strings"
)

// 標籤限制常數（依 AWS 規範）。
const (
	MaxKeyLength       = 128
	MaxValueLength     = 256
	MaxTagsPerResource = 50
)

// 禁止的 key 前綴。
var reservedPrefixes = []string{"aws:"}

// 合法字元正則（key 與 value 皆適用）。
var validTagPattern = regexp.MustCompile(`^[\p{L}\p{N}\s_.:/=+\-@]+$`)

// ValidationError 描述標籤驗證錯誤。
type ValidationError struct {
	Key     string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidateKey 驗證標籤 key。
func ValidateKey(key string) error {
	if key == "" {
		return ValidationError{Key: key, Message: "標籤 key 不可為空"}
	}
	if len(key) > MaxKeyLength {
		return ValidationError{Key: key, Message: "標籤 key 超過 128 字元"}
	}
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(strings.ToLower(key), prefix) {
			return ValidationError{Key: key, Message: "標籤 key 不可使用 aws: 前綴"}
		}
	}
	if !validTagPattern.MatchString(key) {
		return ValidationError{Key: key, Message: "標籤 key 包含無效字元"}
	}
	return nil
}

// ValidateValue 驗證標籤 value。
func ValidateValue(value string) error {
	if len(value) > MaxValueLength {
		return ValidationError{Value: value, Message: "標籤 value 超過 256 字元"}
	}
	if value != "" && !validTagPattern.MatchString(value) {
		return ValidationError{Value: value, Message: "標籤 value 包含無效字元"}
	}
	return nil
}

// ValidateTag 驗證單一標籤。
func ValidateTag(key, value string) error {
	if err := ValidateKey(key); err != nil {
		return err
	}
	return ValidateValue(value)
}

// ValidateTags 驗證標籤集合。
func ValidateTags(tags map[string]string) []error {
	var errs []error
	if len(tags) > MaxTagsPerResource {
		errs = append(errs, errors.New("標籤數量超過 50 個"))
	}
	for k, v := range tags {
		if err := ValidateTag(k, v); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// DetectConflicts 檢查新標籤與現有標籤的衝突。
func DetectConflicts(existing, updates map[string]string) []string {
	var conflicts []string
	for k := range updates {
		if _, exists := existing[k]; exists {
			conflicts = append(conflicts, k)
		}
	}
	return conflicts
}

// MergeTags 合併標籤（updates 覆蓋 existing）。
func MergeTags(existing, updates map[string]string) map[string]string {
	result := make(map[string]string, len(existing)+len(updates))
	for k, v := range existing {
		result[k] = v
	}
	for k, v := range updates {
		result[k] = v
	}
	return result
}
