package config

import (
	"os"
	"strings"
	"testing"
)

// TestResolveValue_PlainValue 测试普通值（不使用环境变量）
func TestResolveValue_PlainValue(t *testing.T) {
	result, err := resolveValue("plain-value-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != "plain-value-123" {
		t.Errorf("Expected 'plain-value-123', got %s", result)
	}
}

// TestResolveValue_EnvVar 测试环境变量解析
func TestResolveValue_EnvVar(t *testing.T) {
	os.Setenv("TEST_VAR", "env-value-456")
	defer os.Unsetenv("TEST_VAR")

	result, err := resolveValue("env:TEST_VAR")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != "env-value-456" {
		t.Errorf("Expected 'env-value-456', got %s", result)
	}
}

// TestResolveValue_EmptyEnvName 测试空环境变量名
func TestResolveValue_EmptyEnvName(t *testing.T) {
	_, err := resolveValue("env:")
	if err == nil {
		t.Error("Expected error for empty env var name, got nil")
	}

	if !strings.Contains(err.Error(), "Empty environment variable name") {
		t.Errorf("Expected error containing 'Empty environment variable name', got: %v", err)
	}
}

// TestResolveValue_EnvNotSet 测试未设置的环境变量
func TestResolveValue_EnvNotSet(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")

	_, err := resolveValue("env:NONEXISTENT_VAR")
	if err == nil {
		t.Error("Expected error for unset env var, got nil")
	}
}

// TestResolveValue_EnvEmpty 测试空的环境变量
func TestResolveValue_EnvEmpty(t *testing.T) {
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	_, err := resolveValue("env:EMPTY_VAR")
	if err == nil {
		t.Error("Expected error for empty env var, got nil")
	}
}

// TestResolveEnvVars_BasicAuth 测试 Basic Auth 环境变量解析
func TestResolveEnvVars_BasicAuth(t *testing.T) {
	os.Setenv("BASIC_PASS", "secret-password")
	os.Setenv("BASIC_PASS_HASH", "$2a$10$examplehash")
	defer os.Unsetenv("BASIC_PASS")
	defer os.Unsetenv("BASIC_PASS_HASH")

	cfg := &Config{
		BasicAuths: []BasicAuthConfig{
			{
				Name:     "test",
				User:     "testuser",
				Pass:     "env:BASIC_PASS",
				PassHash: "env:BASIC_PASS_HASH",
			},
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.BasicAuths[0].Pass != "secret-password" {
		t.Errorf("Expected password 'secret-password', got %s", cfg.BasicAuths[0].Pass)
	}
	if cfg.BasicAuths[0].PassHash != "$2a$10$examplehash" {
		t.Errorf("Expected pass_hash '$2a$10$examplehash', got %s", cfg.BasicAuths[0].PassHash)
	}
}

// TestResolveEnvVars_BearerToken 测试 Bearer Token 环境变量解析
func TestResolveEnvVars_BearerToken(t *testing.T) {
	os.Setenv("BEARER_TOKEN", "token-xyz-789")
	defer os.Unsetenv("BEARER_TOKEN")

	cfg := &Config{
		BearerTokens: []BearerConfig{
			{
				Name:  "service1",
				Token: "env:BEARER_TOKEN",
			},
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.BearerTokens[0].Token != "token-xyz-789" {
		t.Errorf("Expected token 'token-xyz-789', got %s", cfg.BearerTokens[0].Token)
	}
}

// TestResolveEnvVars_APIKey 测试 API Key 环境变量解析
func TestResolveEnvVars_APIKey(t *testing.T) {
	os.Setenv("API_KEY", "apikey-abc-123")
	defer os.Unsetenv("API_KEY")

	cfg := &Config{
		APIKeys: []APIKeyConfig{
			{
				Name: "mobile-app",
				Key:  "env:API_KEY",
			},
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.APIKeys[0].Key != "apikey-abc-123" {
		t.Errorf("Expected key 'apikey-abc-123', got %s", cfg.APIKeys[0].Key)
	}
}

// TestResolveEnvVars_JWTSecret 测试 JWT Secret 环境变量解析
func TestResolveEnvVars_JWTSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "super-secret-jwt-key")
	defer os.Unsetenv("JWT_SECRET")

	cfg := &Config{
		JWT: JWTConfig{
			Secret: "env:JWT_SECRET",
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.JWT.Secret != "super-secret-jwt-key" {
		t.Errorf("Expected secret 'super-secret-jwt-key', got %s", cfg.JWT.Secret)
	}
}

// TestResolveEnvVars_EmptyJWT 测试空的 JWT 配置
func TestResolveEnvVars_EmptyJWT(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "",
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Errorf("Unexpected error for empty JWT: %v", err)
	}
}

// TestResolveEnvVars_MultipleTypes 测试同时解析多种类型
func TestResolveEnvVars_MultipleTypes(t *testing.T) {
	os.Setenv("PASS1", "password1")
	os.Setenv("PASS_HASH1", "$2a$10$hashvalue")
	os.Setenv("TOKEN1", "token1")
	os.Setenv("KEY1", "key1")
	os.Setenv("JWT_SEC", "jwtsecret")
	defer func() {
		os.Unsetenv("PASS1")
		os.Unsetenv("PASS_HASH1")
		os.Unsetenv("TOKEN1")
		os.Unsetenv("KEY1")
		os.Unsetenv("JWT_SEC")
	}()

	cfg := &Config{
		BasicAuths: []BasicAuthConfig{
			{Name: "user1", User: "user1", Pass: "env:PASS1", PassHash: "env:PASS_HASH1"},
		},
		BearerTokens: []BearerConfig{
			{Name: "svc1", Token: "env:TOKEN1"},
		},
		APIKeys: []APIKeyConfig{
			{Name: "app1", Key: "env:KEY1"},
		},
		JWT: JWTConfig{
			Secret: "env:JWT_SEC",
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证所有值都被正确解析
	if cfg.BasicAuths[0].Pass != "password1" {
		t.Errorf("Expected password 'password1', got %s", cfg.BasicAuths[0].Pass)
	}
	if cfg.BasicAuths[0].PassHash != "$2a$10$hashvalue" {
		t.Errorf("Expected pass_hash '$2a$10$hashvalue', got %s", cfg.BasicAuths[0].PassHash)
	}
	if cfg.BearerTokens[0].Token != "token1" {
		t.Errorf("Expected token 'token1', got %s", cfg.BearerTokens[0].Token)
	}
	if cfg.APIKeys[0].Key != "key1" {
		t.Errorf("Expected key 'key1', got %s", cfg.APIKeys[0].Key)
	}
	if cfg.JWT.Secret != "jwtsecret" {
		t.Errorf("Expected secret 'jwtsecret', got %s", cfg.JWT.Secret)
	}
}

// TestResolveEnvVars_ErrorPropagation 测试错误传播
func TestResolveEnvVars_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "Basic Auth with missing env",
			cfg: &Config{
				BasicAuths: []BasicAuthConfig{
					{Name: "test", User: "user", Pass: "env:MISSING_VAR"},
				},
			},
			wantErr: true,
		},
		{
			name: "Basic Auth pass_hash with missing env",
			cfg: &Config{
				BasicAuths: []BasicAuthConfig{
					{Name: "test", User: "user", PassHash: "env:MISSING_VAR"},
				},
			},
			wantErr: true,
		},
		{
			name: "Bearer Token with missing env",
			cfg: &Config{
				BearerTokens: []BearerConfig{
					{Name: "test", Token: "env:MISSING_VAR"},
				},
			},
			wantErr: true,
		},
		{
			name: "API Key with missing env",
			cfg: &Config{
				APIKeys: []APIKeyConfig{
					{Name: "test", Key: "env:MISSING_VAR"},
				},
			},
			wantErr: true,
		},
		{
			name: "JWT with missing env",
			cfg: &Config{
				JWT: JWTConfig{
					Secret: "env:MISSING_VAR",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("MISSING_VAR")

			err := ResolveEnvVars(tt.cfg)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestResolveEnvVars_MixedPlainAndEnv 测试混合使用普通值和环境变量
func TestResolveEnvVars_MixedPlainAndEnv(t *testing.T) {
	os.Setenv("ENV_PASS", "env-password")
	defer os.Unsetenv("ENV_PASS")

	cfg := &Config{
		BasicAuths: []BasicAuthConfig{
			{Name: "user1", User: "user1", Pass: "plain-password"},
			{Name: "user2", User: "user2", Pass: "env:ENV_PASS"},
		},
	}

	err := ResolveEnvVars(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证普通值保持不变
	if cfg.BasicAuths[0].Pass != "plain-password" {
		t.Errorf("Expected 'plain-password', got %s", cfg.BasicAuths[0].Pass)
	}

	// 验证环境变量被解析
	if cfg.BasicAuths[1].Pass != "env-password" {
		t.Errorf("Expected 'env-password', got %s", cfg.BasicAuths[1].Pass)
	}
}
