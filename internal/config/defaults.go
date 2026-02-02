package config

import "os"

const (
	defaultPort         = "8080"
	defaultAuthPath     = "/auth"
	defaultHealthPath   = "/health"
	defaultUserHeader   = "X-Auth-User"
	defaultRoleHeader   = "X-Auth-Role"
	defaultMethodHeader = "X-Auth-Method"
	defaultLogFormat    = "text"
	defaultLogLevel     = "info"
)

// ApplyDefaults 应用默认值到配置
//
//nolint:gocognit,gocyclo // defaulting logic is explicit for clarity
func ApplyDefaults(cfg *Config) {
	// 服务器默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = defaultPort
	}
	if cfg.Server.AuthPath == "" {
		cfg.Server.AuthPath = defaultAuthPath
	}
	if cfg.Server.HealthPath == "" {
		cfg.Server.HealthPath = defaultHealthPath
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 5
	}

	// Header 默认值
	if cfg.Headers.UserHeader == "" {
		cfg.Headers.UserHeader = defaultUserHeader
	}
	if cfg.Headers.RoleHeader == "" {
		cfg.Headers.RoleHeader = defaultRoleHeader
	}
	if cfg.Headers.MethodHeader == "" {
		cfg.Headers.MethodHeader = defaultMethodHeader
	}

	// 日志默认值
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = defaultLogFormat
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = defaultLogLevel
	}

	// 审计日志默认值
	if cfg.Audit.Enabled && cfg.Audit.Output == "" {
		cfg.Audit.Output = "stdout"
	}

	// 速率限制默认值
	if cfg.RateLimit.MaxAttempts == 0 {
		cfg.RateLimit.MaxAttempts = 5 // 默认 5 次尝试
	}
	if cfg.RateLimit.WindowSecs == 0 {
		cfg.RateLimit.WindowSecs = 60 // 默认 60 秒窗口
	}
	if cfg.RateLimit.BanSecs == 0 {
		cfg.RateLimit.BanSecs = 300 // 默认封禁 5 分钟
	}

	// Basic Auth 默认角色
	for i := range cfg.BasicAuths {
		if len(cfg.BasicAuths[i].Roles) == 0 {
			cfg.BasicAuths[i].Roles = []string{"user"}
		}
	}

	// Bearer Token 默认角色
	for i := range cfg.BearerTokens {
		if len(cfg.BearerTokens[i].Roles) == 0 {
			cfg.BearerTokens[i].Roles = []string{"service"}
		}
	}

	// API Key 默认角色
	for i := range cfg.APIKeys {
		if len(cfg.APIKeys[i].Roles) == 0 {
			cfg.APIKeys[i].Roles = []string{"api"}
		}
	}

	// 环境变量覆盖端口
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}
}
