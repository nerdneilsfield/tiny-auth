package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
	apperrors "github.com/nerdneilsfield/tiny-auth/internal/errors"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate configuration file",
		Long:  `Validate the tiny-auth configuration file for syntax and semantic errors.`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runValidate,
	}

	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	// 确定配置文件路径
	cfgPath := configPath
	if len(args) > 0 {
		cfgPath = args[0]
	}

	fmt.Printf("Validating configuration file: %s\n\n", cfgPath)

	// 检查文件是否存在
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		appErr := apperrors.ConfigNotFound(cfgPath)
		fmt.Printf("❌ Error: %s\n", appErr.Message)
		return appErr
	}

	// 加载并验证配置
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("❌ Configuration validation failed:\n")
		fmt.Printf("   %v\n\n", err)
		return err
	}

	// 输出验证结果
	fmt.Println("✅ Configuration is valid")
	fmt.Println()
	printConfigSummary(cfg)

	return nil
}

func printConfigSummary(cfg *config.Config) {
	fmt.Println("📋 Configuration Summary:")
	fmt.Println()

	// 服务器配置
	fmt.Printf("✓ Server:\n")
	fmt.Printf("  - Port: %s\n", cfg.Server.Port)
	fmt.Printf("  - Auth Path: %s\n", cfg.Server.AuthPath)
	fmt.Printf("  - Health Path: %s\n", cfg.Server.HealthPath)
	fmt.Printf("  - Read Timeout: %ds\n", cfg.Server.ReadTimeout)
	fmt.Printf("  - Write Timeout: %ds\n", cfg.Server.WriteTimeout)
	fmt.Println()

	// 日志配置
	fmt.Printf("✓ Logging:\n")
	fmt.Printf("  - Format: %s\n", cfg.Logging.Format)
	fmt.Printf("  - Level: %s\n", cfg.Logging.Level)
	fmt.Println()

	// 认证方法
	if len(cfg.BasicAuths) > 0 {
		fmt.Printf("✓ Basic Auth: %d users configured\n", len(cfg.BasicAuths))
		for _, b := range cfg.BasicAuths {
			fmt.Printf("  - %s (user=%s, roles=%v)\n", b.Name, b.User, b.Roles)
		}
		fmt.Println()
	}

	if len(cfg.BearerTokens) > 0 {
		fmt.Printf("✓ Bearer Tokens: %d tokens configured\n", len(cfg.BearerTokens))
		for _, b := range cfg.BearerTokens {
			fmt.Printf("  - %s (roles=%v)\n", b.Name, b.Roles)
		}
		fmt.Println()
	}

	if len(cfg.APIKeys) > 0 {
		fmt.Printf("✓ API Keys: %d keys configured\n", len(cfg.APIKeys))
		for _, k := range cfg.APIKeys {
			fmt.Printf("  - %s (roles=%v)\n", k.Name, k.Roles)
		}
		fmt.Println()
	}

	if cfg.JWT.Secret != "" {
		fmt.Printf("✓ JWT: enabled\n")
		if cfg.JWT.Issuer != "" {
			fmt.Printf("  - Issuer: %s\n", cfg.JWT.Issuer)
		}
		if cfg.JWT.Audience != "" {
			fmt.Printf("  - Audience: %s\n", cfg.JWT.Audience)
		}
		fmt.Println()
	}

	// 路由策略
	if len(cfg.RoutePolicies) > 0 {
		fmt.Printf("✓ Route Policies: %d policies configured\n", len(cfg.RoutePolicies))
		for i := range cfg.RoutePolicies {
			p := cfg.RoutePolicies[i]
			fmt.Printf("  - %s", p.Name)
			if p.Host != "" {
				fmt.Printf(" (host=%s", p.Host)
				if p.PathPrefix != "" {
					fmt.Printf(", path=%s", p.PathPrefix)
				}
				fmt.Printf(")")
			}
			if p.AllowAnonymous {
				fmt.Printf(" [anonymous]")
			}
			fmt.Println()
		}
		fmt.Println()
	}

	logger.Info("Configuration validated successfully",
		zap.Int("basic_auth", len(cfg.BasicAuths)),
		zap.Int("bearer_tokens", len(cfg.BearerTokens)),
		zap.Int("api_keys", len(cfg.APIKeys)),
		zap.Bool("jwt_enabled", cfg.JWT.Secret != ""),
		zap.Int("policies", len(cfg.RoutePolicies)),
	)
}
