package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
	"github.com/nerdneilsfield/tiny-auth/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the authentication server",
		Long:  `Start the tiny-auth ForwardAuth authentication server.`,
		RunE:  runServer,
	}

	return cmd
}

func runServer(cmd *cobra.Command, args []string) error {
	// 1. 加载配置
	logger.Info("Loading configuration", zap.String("path", configPath))
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
		return err
	}

	// 2. 构建认证存储
	store := auth.BuildStore(cfg)

	// 3. 创建服务器
	srv := server.NewServer(cfg, store, logger)

	// 4. 设置信号处理（优雅关闭 + 配置重载）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// 启动后台协程处理信号
	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGHUP:
				// 配置热重载
				logger.Info("Received SIGHUP, reloading configuration...")
				if err := reloadConfig(srv); err != nil {
					logger.Error("Failed to reload config", zap.Error(err))
				}

			case os.Interrupt, syscall.SIGTERM:
				// 优雅关闭
				logger.Info("Received shutdown signal, shutting down...")
				if err := srv.Shutdown(); err != nil {
					logger.Error("Error during shutdown", zap.Error(err))
				}
				os.Exit(0)
			}
		}
	}()

	// 5. 启动服务器
	logger.Info("Starting server",
		zap.String("port", cfg.Server.Port),
		zap.String("auth_path", cfg.Server.AuthPath),
	)

	if err := srv.Start(); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
		return err
	}

	return nil
}

// reloadConfig 重新加载配置
func reloadConfig(srv *server.Server) error {
	// 重新加载配置文件
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 重新构建认证存储
	store := auth.BuildStore(cfg)

	// 更新服务器配置
	srv.Reload(cfg, store)

	logger.Info("Configuration reloaded successfully",
		zap.Int("basic_auth", len(cfg.BasicAuths)),
		zap.Int("bearer_tokens", len(cfg.BearerTokens)),
		zap.Int("api_keys", len(cfg.APIKeys)),
		zap.Int("policies", len(cfg.RoutePolicies)),
	)

	return nil
}
