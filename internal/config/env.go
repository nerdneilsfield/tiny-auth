package config

import (
	"fmt"
	"os"
	"strings"
)

// ResolveEnvVars 解析配置中的环境变量
// 支持语法: "env:VAR_NAME"
func ResolveEnvVars(cfg *Config) error {
	// 解析 Basic Auth 密码
	for i := range cfg.BasicAuths {
		resolved, err := resolveValue(cfg.BasicAuths[i].Pass)
		if err != nil {
			return fmt.Errorf("basic_auth[%s].pass: %w", cfg.BasicAuths[i].Name, err)
		}
		cfg.BasicAuths[i].Pass = resolved
	}

	// 解析 Bearer Token
	for i := range cfg.BearerTokens {
		resolved, err := resolveValue(cfg.BearerTokens[i].Token)
		if err != nil {
			return fmt.Errorf("bearer_token[%s].token: %w", cfg.BearerTokens[i].Name, err)
		}
		cfg.BearerTokens[i].Token = resolved
	}

	// 解析 API Key
	for i := range cfg.APIKeys {
		resolved, err := resolveValue(cfg.APIKeys[i].Key)
		if err != nil {
			return fmt.Errorf("api_key[%s].key: %w", cfg.APIKeys[i].Name, err)
		}
		cfg.APIKeys[i].Key = resolved
	}

	// 解析 JWT Secret
	if cfg.JWT.Secret != "" {
		resolved, err := resolveValue(cfg.JWT.Secret)
		if err != nil {
			return fmt.Errorf("jwt.secret: %w", err)
		}
		cfg.JWT.Secret = resolved
	}

	return nil
}

// resolveValue 解析单个值的环境变量
// 如果值以 "env:" 开头，则从环境变量读取
func resolveValue(value string) (string, error) {
	if !strings.HasPrefix(value, "env:") {
		return value, nil
	}

	envVar := strings.TrimPrefix(value, "env:")
	if envVar == "" {
		return "", fmt.Errorf("empty environment variable name")
	}

	envValue := os.Getenv(envVar)
	if envValue == "" {
		return "", fmt.Errorf("environment variable %q is not set or empty", envVar)
	}

	return envValue, nil
}
