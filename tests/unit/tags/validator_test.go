// Package tags 提供標籤驗證功能的單元測試。
package tags

import (
	"strings"
	"testing"

	"github.com/vin/ck123gogo/internal/tags"
)

func TestValidateKey_Empty(t *testing.T) {
	err := tags.ValidateKey("")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestValidateKey_TooLong(t *testing.T) {
	longKey := strings.Repeat("a", 129)
	err := tags.ValidateKey(longKey)
	if err == nil {
		t.Error("expected error for key > 128 chars")
	}
}

func TestValidateKey_ReservedPrefix(t *testing.T) {
	err := tags.ValidateKey("aws:reserved")
	if err == nil {
		t.Error("expected error for aws: prefix")
	}
}

func TestValidateKey_Valid(t *testing.T) {
	validKeys := []string{"Name", "Environment", "cost-center", "Project_123"}
	for _, k := range validKeys {
		if err := tags.ValidateKey(k); err != nil {
			t.Errorf("unexpected error for key %q: %v", k, err)
		}
	}
}

func TestValidateValue_TooLong(t *testing.T) {
	longValue := strings.Repeat("v", 257)
	err := tags.ValidateValue(longValue)
	if err == nil {
		t.Error("expected error for value > 256 chars")
	}
}

func TestValidateValue_Valid(t *testing.T) {
	validValues := []string{"", "production", "dev-123", "test_env"}
	for _, v := range validValues {
		if err := tags.ValidateValue(v); err != nil {
			t.Errorf("unexpected error for value %q: %v", v, err)
		}
	}
}

func TestValidateTags_TooMany(t *testing.T) {
	manyTags := make(map[string]string)
	for i := 0; i < 51; i++ {
		manyTags[strings.Repeat("k", i+1)] = "v"
	}
	errs := tags.ValidateTags(manyTags)
	if len(errs) == 0 {
		t.Error("expected error for > 50 tags")
	}
}

func TestDetectConflicts(t *testing.T) {
	existing := map[string]string{"Name": "old", "Env": "prod"}
	updates := map[string]string{"Name": "new", "Team": "dev"}

	conflicts := tags.DetectConflicts(existing, updates)
	if len(conflicts) != 1 || conflicts[0] != "Name" {
		t.Errorf("expected conflict on 'Name', got %v", conflicts)
	}
}

func TestMergeTags(t *testing.T) {
	existing := map[string]string{"Name": "old", "Env": "prod"}
	updates := map[string]string{"Name": "new", "Team": "dev"}

	merged := tags.MergeTags(existing, updates)
	if merged["Name"] != "new" {
		t.Error("expected Name to be overwritten")
	}
	if merged["Env"] != "prod" {
		t.Error("expected Env to be preserved")
	}
	if merged["Team"] != "dev" {
		t.Error("expected Team to be added")
	}
}
