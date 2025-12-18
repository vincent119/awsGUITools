package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/vincent119/awsGUITools/internal/app"
)

var (
	version    = "dev"
	configPath string
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "aws-tui",
		Short:        "以終端 GUI 巡檢與管理 AWS 資源的工具",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			application, err := app.New(
				app.WithVersion(version),
				app.WithConfigPath(configPath),
			)
			if err != nil {
				return err
			}

			return application.Run(ctx)
		},
	}

	cmd.Version = version
	cmd.Flags().StringVar(&configPath, "config", "", "指定組態檔路徑（預設依環境變數載入）")

	return cmd
}
