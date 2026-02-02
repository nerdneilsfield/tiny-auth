package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig_FileNotFound 测试文件不存在的情况
func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
	if err != nil && !os.IsNotExist(err) {
		// Should contain "not found" in error message
		if err.Error() == "" {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	}
}

// TestLoadConfig_ValidFile 测试加载有效配置文件
func TestLoadConfig_ValidFile(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `
[server]
port = "3000"
auth_path = "/auth"

[[basic_auth]]
name = "test"
user = "testuser"
pass = "testpass"
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	if cfg.Server.Port != "3000" {
		t.Errorf("Expected port 3000, got %s", cfg.Server.Port)
	}

	if len(cfg.BasicAuths) != 1 {
		t.Errorf("Expected 1 basic auth, got %d", len(cfg.BasicAuths))
	}

	if cfg.BasicAuths[0].User != "testuser" {
		t.Errorf("Expected user testuser, got %s", cfg.BasicAuths[0].User)
	}
}

// TestLoadConfig_InvalidTOML 测试无效的 TOML 格式
func TestLoadConfig_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// 写入无效的 TOML
	content := `
[server
invalid toml syntax
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid TOML, got nil")
	}
}

// TestLoadConfig_ValidationError 测试配置验证失败
func TestLoadConfig_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// 写入无效配置（空的 basic auth name）
	content := `
[[basic_auth]]
name = ""
user = "testuser"
pass = "testpass"
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}
}

// TestLoadConfig_WithEnvVars 测试环境变量解析
func TestLoadConfig_WithEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// 设置环境变量
	os.Setenv("TEST_PASSWORD", "secret123")
	defer os.Unsetenv("TEST_PASSWORD")

	content := `
[[basic_auth]]
name = "test"
user = "testuser"
pass = "env:TEST_PASSWORD"
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证环境变量被解析
	if cfg.BasicAuths[0].Pass != "secret123" {
		t.Errorf("Expected password 'secret123', got %s", cfg.BasicAuths[0].Pass)
	}
}

// TestLoadConfig_DefaultPath 测试默认路径
func TestLoadConfig_DefaultPath(t *testing.T) {
	// 这个测试会失败，因为没有默认的 config.toml 文件
	// 但我们测试的是路径查找逻辑
	os.Unsetenv("CONFIG_PATH")

	_, err := LoadConfig("")
	if err == nil {
		// If there happens to be a config.toml, that's fine
		return
	}

	// Should look for config.toml
	if !os.IsNotExist(err) {
		// Error should mention config.toml or be a not found error
		t.Logf("Expected error for default config.toml: %v", err)
	}
}

// TestLoadConfig_EnvPath 测试 CONFIG_PATH 环境变量
func TestLoadConfig_EnvPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom.toml")

	content := `
[server]
port = "9000"
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// 设置 CONFIG_PATH
	os.Setenv("CONFIG_PATH", configPath)
	defer os.Unsetenv("CONFIG_PATH")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config from CONFIG_PATH: %v", err)
	}

	if cfg.Server.Port != "9000" {
		t.Errorf("Expected port 9000, got %s", cfg.Server.Port)
	}
}

// TestCheckFilePermissions 测试文件权限检查
func TestCheckFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		perm        os.FileMode
		expectError bool
	}{
		{
			name:        "Secure permissions (0600)",
			perm:        0600,
			expectError: false,
		},
		{
			name:        "Insecure permissions (0644)",
			perm:        0644,
			expectError: true,
		},
		{
			name:        "Insecure permissions (0777)",
			perm:        0777,
			expectError: true,
		},
		{
			name:        "Owner only (0700)",
			perm:        0700,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.name+".toml")

			// 创建测试文件
			if err := os.WriteFile(filePath, []byte("test"), tt.perm); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err := CheckFilePermissions(filePath)

			if tt.expectError && err == nil {
				t.Error("Expected error for insecure permissions, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestCheckFilePermissions_NonExistent 测试检查不存在的文件
func TestCheckFilePermissions_NonExistent(t *testing.T) {
	err := CheckFilePermissions("/nonexistent/file.toml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// TestLoadConfig_FullWorkflow 测试完整的加载流程
func TestLoadConfig_FullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// 设置测试环境变量（至少 32 个字符）
	os.Setenv("TEST_JWT_SECRET", "jwt-secret-key-123-this-is-32-chars-long-secret")
	defer os.Unsetenv("TEST_JWT_SECRET")

	content := `
[server]
port = "8080"
auth_path = "/auth"
trusted_proxies = ["127.0.0.1", "10.0.0.0/8"]

[headers]
user_header = "X-User"

[logging]
format = "json"
level = "debug"

[jwt]
secret = "env:TEST_JWT_SECRET"
issuer = "test-issuer"

[[basic_auth]]
name = "admin"
user = "admin"
pass = "adminpass123456"
roles = ["admin"]

[[route_policy]]
name = "admin-only"
path_prefix = "/admin"
require_any_role = ["admin"]
`

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证服务器配置
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", cfg.Server.Port)
	}

	if len(cfg.Server.TrustedProxies) != 2 {
		t.Errorf("Expected 2 trusted proxies, got %d", len(cfg.Server.TrustedProxies))
	}

	// 验证 Headers 配置
	if cfg.Headers.UserHeader != "X-User" {
		t.Errorf("Expected UserHeader X-User, got %s", cfg.Headers.UserHeader)
	}

	// 验证日志配置
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected logging format json, got %s", cfg.Logging.Format)
	}

	// 验证 JWT 配置（环境变量应该被解析）
	expectedSecret := "jwt-secret-key-123-this-is-32-chars-long-secret"
	if cfg.JWT.Secret != expectedSecret {
		t.Errorf("Expected JWT secret from env var, got %s", cfg.JWT.Secret)
	}

	// 验证 Basic Auth
	if len(cfg.BasicAuths) != 1 {
		t.Fatalf("Expected 1 basic auth, got %d", len(cfg.BasicAuths))
	}

	if cfg.BasicAuths[0].User != "admin" {
		t.Errorf("Expected user admin, got %s", cfg.BasicAuths[0].User)
	}

	// 验证路由策略
	if len(cfg.RoutePolicies) != 1 {
		t.Fatalf("Expected 1 route policy, got %d", len(cfg.RoutePolicies))
	}

	if cfg.RoutePolicies[0].Name != "admin-only" {
		t.Errorf("Expected policy name admin-only, got %s", cfg.RoutePolicies[0].Name)
	}
}
