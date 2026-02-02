package config

import "os"

// ApplyDefaults 应用默认值到配置
func ApplyDefaults(cfg *Config) {
	// 服务器默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.AuthPath == "" {
		cfg.Server.AuthPath = "/auth"
	}
	if cfg.Server.HealthPath == "" {
		cfg.Server.HealthPath = "/health"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 5
	}

	// Header 默认值
	if cfg.Headers.UserHeader == "" {
		cfg.Headers.UserHeader = "X-Auth-User"
	}
	if cfg.Headers.RoleHeader == "" {
		cfg.Headers.RoleHeader = "X-Auth-Role"
	}
	if cfg.Headers.MethodHeader == "" {
		cfg.Headers.MethodHeader = "X-Auth-Method"
	}

	// 日志默认值
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "text"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
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
