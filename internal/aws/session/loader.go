// Package session 提供 AWS 組態載入與快取邏輯。
package session

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

// Loader 定義載入 AWS 組態的介面，便於測試與替換。
type Loader interface {
	Config(ctx context.Context, profile, region string) (aws.Config, error)
}

// CachedLoader 會快取同 profile + region 的 aws.Config，避免重複解析。
type CachedLoader struct {
	mu    sync.RWMutex
	cache map[string]aws.Config
}

// NewLoader 建立具快取能力的 Loader。
func NewLoader() *CachedLoader {
	return &CachedLoader{
		cache: make(map[string]aws.Config),
	}
}

// Config 載入指定 profile 與 region 的 aws.Config，並快取結果。
func (l *CachedLoader) Config(ctx context.Context, profile, region string) (aws.Config, error) {
	key := fmt.Sprintf("%s|%s", profile, region)

	l.mu.RLock()
	if cfg, ok := l.cache[key]; ok {
		l.mu.RUnlock()
		return cfg, nil
	}
	l.mu.RUnlock()

	var (
		cfg aws.Config
		err error
	)
	switch {
	case profile != "" && region != "":
		cfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithSharedConfigProfile(profile),
			awsconfig.WithRegion(region))
	case profile != "":
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithSharedConfigProfile(profile))
	case region != "":
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	default:
		cfg, err = awsconfig.LoadDefaultConfig(ctx)
	}
	if err != nil {
		return aws.Config{}, fmt.Errorf("load aws config (profile=%s, region=%s): %w", profile, region, err)
	}

	l.mu.Lock()
	l.cache[key] = cfg
	l.mu.Unlock()

	return cfg, nil
}
