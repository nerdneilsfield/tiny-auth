package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	verbose    bool
	configPath string
	logger     *zap.Logger
)

func init() {
	// 初始化 logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
}

func newRootCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tiny-auth",
		Short: "tiny-auth - A lightweight authentication service for Traefik ForwardAuth",
		Long: `tiny-auth is a high-performance authentication service designed for Traefik ForwardAuth middleware.
It supports multiple authentication methods (Basic Auth, Bearer Token, API Key, JWT) with flexible route-based policies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				// 切换到开发模式的 logger（更详细）
				var err error
				logger, err = zap.NewDevelopment()
				if err != nil {
					panic(fmt.Sprintf("failed to create development logger: %v", err))
				}
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.toml", "Config file path")

	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newVersionCmd(version, buildTime, gitCommit))
	cmd.AddCommand(newHashPasswordCmd())

	return cmd
}

func Execute(version string, buildTime string, gitCommit string) error {
	defer func() {
		_ = logger.Sync()
	}()

	if err := newRootCmd(version, buildTime, gitCommit).Execute(); err != nil {
		logger.Fatal("error executing root command", zap.Error(err))
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
