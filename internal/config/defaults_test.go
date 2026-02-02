package config

import (
	"os"
	"testing"
)

// TestApplyDefaults_ServerDefaults 测试服务器默认值
func TestApplyDefaults_ServerDefaults(t *testing.T) {
	cfg := &Config{}

	ApplyDefaults(cfg)

	// 验证服务器默认值
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Server.AuthPath != "/auth" {
		t.Errorf("Expected default auth_path /auth, got %s", cfg.Server.AuthPath)
	}

	if cfg.Server.HealthPath != "/health" {
		t.Errorf("Expected default health_path /health, got %s", cfg.Server.HealthPath)
	}

	if cfg.Server.ReadTimeout != 5 {
		t.Errorf("Expected default read_timeout 5, got %d", cfg.Server.ReadTimeout)
	}

	if cfg.Server.WriteTimeout != 5 {
		t.Errorf("Expected default write_timeout 5, got %d", cfg.Server.WriteTimeout)
	}
}

// TestApplyDefaults_ServerCustomValues 测试自定义值不被覆盖
func TestApplyDefaults_ServerCustomValues(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         "9000",
			AuthPath:     "/custom-auth",
			HealthPath:   "/custom-health",
			ReadTimeout:  10,
			WriteTimeout: 15,
		},
	}

	ApplyDefaults(cfg)

	// 验证自定义值不被覆盖
	if cfg.Server.Port != "9000" {
		t.Errorf("Custom port should not be overridden, got %s", cfg.Server.Port)
	}

	if cfg.Server.AuthPath != "/custom-auth" {
		t.Errorf("Custom auth_path should not be overridden, got %s", cfg.Server.AuthPath)
	}

	if cfg.Server.ReadTimeout != 10 {
		t.Errorf("Custom read_timeout should not be overridden, got %d", cfg.Server.ReadTimeout)
	}
}

// TestApplyDefaults_HeaderDefaults 测试 Header 默认值
func TestApplyDefaults_HeaderDefaults(t *testing.T) {
	cfg := &Config{}

	ApplyDefaults(cfg)

	if cfg.Headers.UserHeader != "X-Auth-User" {
		t.Errorf("Expected default UserHeader X-Auth-User, got %s", cfg.Headers.UserHeader)
	}

	if cfg.Headers.RoleHeader != "X-Auth-Role" {
		t.Errorf("Expected default RoleHeader X-Auth-Role, got %s", cfg.Headers.RoleHeader)
	}

	if cfg.Headers.MethodHeader != "X-Auth-Method" {
		t.Errorf("Expected default MethodHeader X-Auth-Method, got %s", cfg.Headers.MethodHeader)
	}
}

// TestApplyDefaults_LoggingDefaults 测试日志默认值
func TestApplyDefaults_LoggingDefaults(t *testing.T) {
	cfg := &Config{}

	ApplyDefaults(cfg)

	if cfg.Logging.Format != "text" {
		t.Errorf("Expected default logging format text, got %s", cfg.Logging.Format)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default logging level info, got %s", cfg.Logging.Level)
	}
}

// TestApplyDefaults_BasicAuthRoles 测试 Basic Auth 默认角色
func TestApplyDefaults_BasicAuthRoles(t *testing.T) {
	cfg := &Config{
		BasicAuths: []BasicAuthConfig{
			{Name: "user1", User: "user1", Pass: "pass1"},
			{Name: "user2", User: "user2", Pass: "pass2", Roles: []string{"admin"}},
		},
	}

	ApplyDefaults(cfg)

	// 验证第一个用户获得默认角色
	if len(cfg.BasicAuths[0].Roles) != 1 || cfg.BasicAuths[0].Roles[0] != "user" {
		t.Errorf("Expected default role 'user', got %v", cfg.BasicAuths[0].Roles)
	}

	// 验证第二个用户的自定义角色不被覆盖
	if len(cfg.BasicAuths[1].Roles) != 1 || cfg.BasicAuths[1].Roles[0] != "admin" {
		t.Errorf("Custom roles should not be overridden, got %v", cfg.BasicAuths[1].Roles)
	}
}

// TestApplyDefaults_BearerTokenRoles 测试 Bearer Token 默认角色
func TestApplyDefaults_BearerTokenRoles(t *testing.T) {
	cfg := &Config{
		BearerTokens: []BearerConfig{
			{Name: "svc1", Token: "token1"},
			{Name: "svc2", Token: "token2", Roles: []string{"api"}},
		},
	}

	ApplyDefaults(cfg)

	// 验证第一个 token 获得默认角色
	if len(cfg.BearerTokens[0].Roles) != 1 || cfg.BearerTokens[0].Roles[0] != "service" {
		t.Errorf("Expected default role 'service', got %v", cfg.BearerTokens[0].Roles)
	}

	// 验证第二个 token 的自定义角色不被覆盖
	if len(cfg.BearerTokens[1].Roles) != 1 || cfg.BearerTokens[1].Roles[0] != "api" {
		t.Errorf("Custom roles should not be overridden, got %v", cfg.BearerTokens[1].Roles)
	}
}

// TestApplyDefaults_APIKeyRoles 测试 API Key 默认角色
func TestApplyDefaults_APIKeyRoles(t *testing.T) {
	cfg := &Config{
		APIKeys: []APIKeyConfig{
			{Name: "app1", Key: "key1"},
			{Name: "app2", Key: "key2", Roles: []string{"mobile"}},
		},
	}

	ApplyDefaults(cfg)

	// 验证第一个 key 获得默认角色
	if len(cfg.APIKeys[0].Roles) != 1 || cfg.APIKeys[0].Roles[0] != "api" {
		t.Errorf("Expected default role 'api', got %v", cfg.APIKeys[0].Roles)
	}

	// 验证第二个 key 的自定义角色不被覆盖
	if len(cfg.APIKeys[1].Roles) != 1 || cfg.APIKeys[1].Roles[0] != "mobile" {
		t.Errorf("Custom roles should not be overridden, got %v", cfg.APIKeys[1].Roles)
	}
}

// TestApplyDefaults_PortEnvVar 测试 PORT 环境变量覆盖
func TestApplyDefaults_PortEnvVar(t *testing.T) {
	// 保存原始 PORT 值
	originalPort := os.Getenv("PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// 设置 PORT 环境变量
	os.Setenv("PORT", "7777")

	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},
	}

	ApplyDefaults(cfg)

	// 验证 PORT 环境变量覆盖配置值
	if cfg.Server.Port != "7777" {
		t.Errorf("Expected PORT env var to override config, got %s", cfg.Server.Port)
	}
}

// TestApplyDefaults_EmptyPortWithEnv 测试空端口配置使用环境变量
func TestApplyDefaults_EmptyPortWithEnv(t *testing.T) {
	originalPort := os.Getenv("PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	os.Setenv("PORT", "3333")

	cfg := &Config{}

	ApplyDefaults(cfg)

	// 应该先应用默认值 8080，然后被环境变量 3333 覆盖
	if cfg.Server.Port != "3333" {
		t.Errorf("Expected PORT env var, got %s", cfg.Server.Port)
	}
}

// TestApplyDefaults_NoPortEnv 测试没有 PORT 环境变量
func TestApplyDefaults_NoPortEnv(t *testing.T) {
	originalPort := os.Getenv("PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	os.Unsetenv("PORT")

	cfg := &Config{}

	ApplyDefaults(cfg)

	// 应该使用默认端口
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}
}

// TestApplyDefaults_FullConfig 测试完整配置的默认值应用
func TestApplyDefaults_FullConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "", // 空值，应该使用默认值
		},
		BasicAuths: []BasicAuthConfig{
			{Name: "user1", User: "user1", Pass: "pass1"}, // 无角色
		},
		BearerTokens: []BearerConfig{
			{Name: "svc1", Token: "token1"}, // 无角色
		},
		APIKeys: []APIKeyConfig{
			{Name: "app1", Key: "key1"}, // 无角色
		},
	}

	ApplyDefaults(cfg)

	// 验证所有默认值都被正确应用
	if cfg.Server.Port != "8080" {
		t.Errorf("Server port default not applied")
	}

	if cfg.Server.AuthPath != "/auth" {
		t.Errorf("Auth path default not applied")
	}

	if cfg.Headers.UserHeader != "X-Auth-User" {
		t.Errorf("UserHeader default not applied")
	}

	if cfg.Logging.Format != "text" {
		t.Errorf("Logging format default not applied")
	}

	if len(cfg.BasicAuths[0].Roles) == 0 || cfg.BasicAuths[0].Roles[0] != "user" {
		t.Errorf("Basic auth role default not applied")
	}

	if len(cfg.BearerTokens[0].Roles) == 0 || cfg.BearerTokens[0].Roles[0] != "service" {
		t.Errorf("Bearer token role default not applied")
	}

	if len(cfg.APIKeys[0].Roles) == 0 || cfg.APIKeys[0].Roles[0] != "api" {
		t.Errorf("API key role default not applied")
	}
}

// TestApplyDefaults_EmptyRolesSlice 测试空角色切片
func TestApplyDefaults_EmptyRolesSlice(t *testing.T) {
	cfg := &Config{
		BasicAuths: []BasicAuthConfig{
			{Name: "user1", User: "user1", Pass: "pass1", Roles: []string{}},
		},
	}

	ApplyDefaults(cfg)

	// 空切片应该被视为"没有角色"，需要添加默认角色
	if len(cfg.BasicAuths[0].Roles) != 1 || cfg.BasicAuths[0].Roles[0] != "user" {
		t.Errorf("Empty roles slice should be filled with default, got %v", cfg.BasicAuths[0].Roles)
	}
}

// TestApplyDefaults_Idempotent 测试幂等性（多次调用应该产生相同结果）
func TestApplyDefaults_Idempotent(t *testing.T) {
	cfg := &Config{}

	ApplyDefaults(cfg)
	firstPort := cfg.Server.Port

	ApplyDefaults(cfg)
	secondPort := cfg.Server.Port

	if firstPort != secondPort {
		t.Errorf("ApplyDefaults is not idempotent: first=%s, second=%s", firstPort, secondPort)
	}
}
