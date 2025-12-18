// Package app 提供應用程式生命週期管理與主要業務邏輯。
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/vincent119/awsGUITools/internal/app/config"
	"github.com/vincent119/awsGUITools/internal/app/state"
	awsclients "github.com/vincent119/awsGUITools/internal/aws/clients"
	"github.com/vincent119/awsGUITools/internal/aws/session"
	"github.com/vincent119/awsGUITools/internal/i18n"
	"github.com/vincent119/awsGUITools/internal/observability"
	"github.com/vincent119/awsGUITools/internal/service/resource"
	"github.com/vincent119/awsGUITools/internal/theme"
	"github.com/vincent119/awsGUITools/internal/ui"
)

// App 代表整體 CLI 應用生命週期。
type App struct {
	cfg      config.Config
	cfgPath  string
	logger   *slog.Logger
	version  string
	started  time.Time
	shutdown chan struct{}

	stateStore *state.Store
	themeMgr   *theme.Manager
	uiRoot     *ui.Root

	sessionLoader session.Loader
	clientFactory *awsclients.Factory
	metrics       *observability.AWSCallMetrics
	resources     *resource.Service
}

// Option 允許在建立 App 時注入額外設定。
type Option func(*App)

// WithConfigPath 設定組態檔路徑。
func WithConfigPath(path string) Option {
	return func(a *App) {
		a.cfgPath = path
	}
}

// WithVersion 設定應用程式版本。
func WithVersion(v string) Option {
	return func(a *App) {
		if v != "" {
			a.version = v
		}
	}
}

// New 建立 App 實例並載入設定。
func New(opts ...Option) (*App, error) {
	a := &App{
		version:  "dev",
		logger:   observability.NewLogger(),
		shutdown: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(a)
	}

	cfg, err := config.Load(a.cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	a.cfg = cfg

	// 初始化 i18n（根據設定的語言）
	i18n.SetLanguage(i18n.Language(cfg.Language))

	// 使用 NewWithProfiles 以支援 profile 選擇與自動 region 切換
	a.stateStore = state.NewWithProfiles(cfg.Profile, cfg.Region, cfg.Theme, cfg.Language, cfg.Profiles)

	themeMgr, err := theme.NewManager()
	if err != nil {
		return nil, fmt.Errorf("init theme manager: %w", err)
	}
	a.themeMgr = themeMgr

	// 先初始化 AWS 相關元件（順序重要！）
	a.sessionLoader = session.NewLoader()
	a.clientFactory = awsclients.NewFactory(a.sessionLoader)
	a.metrics = observability.NewAWSCallMetrics(a.logger)

	// 現在才能建立 resource service（依賴 clientFactory）
	a.resources = resource.NewService(a.clientFactory, a.metrics, cfg.RequestTimeout, a.stateStore)

	uiRoot, err := ui.NewRoot(cfg, themeMgr, a.stateStore, a.resources)
	if err != nil {
		return nil, fmt.Errorf("init ui root: %w", err)
	}
	a.uiRoot = uiRoot

	return a, nil
}

// Run 啟動應用主流程（後續將串接 UI）。
func (a *App) Run(ctx context.Context) error {
	a.started = time.Now()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a.logger.Info("aws-tui 啟動",
		slog.String("version", a.version),
		slog.String("profile", a.cfg.Profile),
		slog.String("region", a.cfg.Region),
		slog.String("theme", a.cfg.Theme),
	)

	if a.uiRoot != nil {
		if err := a.uiRoot.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("run ui: %w", err)
		}
	} else {
		<-ctx.Done()
		a.logger.Info("收到終止訊號，準備結束", slog.String("reason", ctx.Err().Error()))
	}

	a.logger.Info("aws-tui 結束", slog.Duration("uptime", time.Since(a.started)))

	return nil
}
