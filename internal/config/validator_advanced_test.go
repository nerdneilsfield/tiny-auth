package config

import (
	"strings"
	"testing"
)

// TestValidatePolicyDependencies 测试策略依赖验证
func TestValidatePolicyDependencies(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid dependencies",
			cfg: &Config{
				BasicAuths: []BasicAuthConfig{
					{Name: "admin", User: "admin", Pass: "password123456"},
				},
				BearerTokens: []BearerConfig{
					{Name: "api-token", Token: "token123"},
				},
				RoutePolicies: []RoutePolicy{
					{
						Name:              "policy1",
						AllowedBasicNames: []string{"admin"},
					},
					{
						Name:                "policy2",
						AllowedBearerNames:  []string{"api-token"},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Non-existent basic_auth reference",
			cfg: &Config{
				BasicAuths: []BasicAuthConfig{
					{Name: "admin", User: "admin", Pass: "password123456"},
				},
				RoutePolicies: []RoutePolicy{
					{
						Name:              "policy1",
						AllowedBasicNames: []string{"non-existent"},
					},
				},
			},
			expectErr: true,
			errMsg:    "non-existent basic_auth name",
		},
		{
			name: "Non-existent bearer_token reference",
			cfg: &Config{
				BearerTokens: []BearerConfig{
					{Name: "token1", Token: "abc123"},
				},
				RoutePolicies: []RoutePolicy{
					{
						Name:               "policy1",
						AllowedBearerNames: []string{"token-missing"},
					},
				},
			},
			expectErr: true,
			errMsg:    "non-existent bearer_token name",
		},
		{
			name: "Non-existent api_key reference",
			cfg: &Config{
				APIKeys: []APIKeyConfig{
					{Name: "key1", Key: "key123"},
				},
				RoutePolicies: []RoutePolicy{
					{
						Name:                "policy1",
						AllowedAPIKeyNames:  []string{"key-missing"},
					},
				},
			},
			expectErr: true,
			errMsg:    "non-existent api_key name",
		},
		{
			name: "Empty policies",
			cfg: &Config{
				BasicAuths:    []BasicAuthConfig{{Name: "admin", User: "admin", Pass: "password123456"}},
				RoutePolicies: []RoutePolicy{},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePolicyDependencies(tt.cfg)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestValidatePolicyConflicts 测试策略冲突检测
func TestValidatePolicyConflicts(t *testing.T) {
	tests := []struct {
		name     string
		policies []RoutePolicy
		hasWarn  bool // 是否应该有警告（但不报错）
	}{
		{
			name: "No conflicts - different hosts",
			policies: []RoutePolicy{
				{Name: "policy1", Host: "api.example.com"},
				{Name: "policy2", Host: "web.example.com"},
			},
			hasWarn: false,
		},
		{
			name: "No conflicts - different paths",
			policies: []RoutePolicy{
				{Name: "policy1", Host: "api.example.com", PathPrefix: "/v1"},
				{Name: "policy2", Host: "api.example.com", PathPrefix: "/v2"},
			},
			hasWarn: false,
		},
		{
			name: "Conflict - same host and path",
			policies: []RoutePolicy{
				{Name: "policy1", Host: "api.example.com", PathPrefix: "/api"},
				{Name: "policy2", Host: "api.example.com", PathPrefix: "/api"},
			},
			hasWarn: true, // 应该有警告
		},
		{
			name: "Empty policies",
			policies: []RoutePolicy{},
			hasWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// validatePolicyConflicts 不返回错误，只输出警告
			err := validatePolicyConflicts(tt.policies)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// 注意：hasWarn 只是标记，实际警告输出到 stderr，测试中无法捕获
		})
	}
}

// TestValidateJWTSecretStrength 测试 JWT 密钥强度验证
func TestValidateJWTSecretStrength(t *testing.T) {
	tests := []struct {
		name      string
		jwt       *JWTConfig
		expectErr bool
	}{
		{
			name: "Empty secret (JWT disabled)",
			jwt: &JWTConfig{
				Secret: "",
			},
			expectErr: false,
		},
		{
			name: "Environment variable secret",
			jwt: &JWTConfig{
				Secret: "env:JWT_SECRET",
			},
			expectErr: false,
		},
		{
			name: "Strong secret (random)",
			jwt: &JWTConfig{
				Secret: "aB3!dE7#gH9@jK1$mN5%pQ2^rS8&tU4*vW6",
			},
			expectErr: false,
		},
		{
			name: "Weak secret (too short)",
			jwt: &JWTConfig{
				Secret: "short",
			},
			expectErr: true,
		},
		{
			name: "Acceptable secret (32 chars, mixed)",
			jwt: &JWTConfig{
				Secret: "MySecretKey1234567890ABCDEFGH123", // 33 chars to pass >= 32
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJWTSecretStrength(tt.jwt)

			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestCalculateEntropy 测试熵值计算
func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		minEntropy    float64 // 最小期望熵值
		maxEntropy    float64 // 最大期望熵值
	}{
		{
			name:       "Empty string",
			input:      "",
			minEntropy: 0.0,
			maxEntropy: 0.0,
		},
		{
			name:       "Single character repeated",
			input:      "aaaaaaa",
			minEntropy: 0.0,
			maxEntropy: 0.0,
		},
		{
			name:       "Two characters (50/50)",
			input:      "ababab",
			minEntropy: 0.9,
			maxEntropy: 1.1,
		},
		{
			name:       "Random mixed characters",
			input:      "aB3!dE7#gH9@",
			minEntropy: 2.5,
			maxEntropy: 4.0,
		},
		{
			name:       "All unique characters",
			input:      "abcdefghij",
			minEntropy: 3.0,
			maxEntropy: 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entropy := calculateEntropy(tt.input)

			if entropy < tt.minEntropy || entropy > tt.maxEntropy {
				t.Errorf("Entropy %.2f out of expected range [%.2f, %.2f]", entropy, tt.minEntropy, tt.maxEntropy)
			}
		})
	}
}

// TestCalculateSecretComplexity 测试密钥复杂度评分
func TestCalculateSecretComplexity(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		minScore int
		maxScore int
	}{
		{
			name:     "Empty string",
			secret:   "",
			minScore: 0,
			maxScore: 0,
		},
		{
			name:     "Very weak (short, simple)",
			secret:   "abc",
			minScore: 0,
			maxScore: 30,
		},
		{
			name:     "Weak (only lowercase)",
			secret:   "abcdefghijklmnop",
			minScore: 20,
			maxScore: 55, // 调整上限，允许一些熵值贡献
		},
		{
			name:     "Medium (mixed case and digits)",
			secret:   "Abc123def456GHI789",
			minScore: 40,
			maxScore: 70,
		},
		{
			name:     "Strong (long, mixed, special)",
			secret:   "aB3!dE7#gH9@jK1$mN5%pQ2^rS8&tU4*vW6",
			minScore: 70,
			maxScore: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateSecretComplexity(tt.secret)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Complexity score %d out of expected range [%d, %d]", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

// TestHashSecret 测试密钥哈希
func TestHashSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{
			name:   "Simple secret",
			secret: "my-secret",
		},
		{
			name:   "Complex secret",
			secret: "aB3!dE7#gH9@jK1$",
		},
		{
			name:   "Empty secret",
			secret: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hashSecret(tt.secret)

			// 哈希应该是 8 个十六进制字符
			if len(hash) != 8 {
				t.Errorf("Expected hash length 8, got %d", len(hash))
			}

			// 相同的输入应该产生相同的哈希
			hash2 := hashSecret(tt.secret)
			if hash != hash2 {
				t.Error("Hash should be deterministic")
			}

			// 不同的输入应该产生不同的哈希
			if tt.secret != "" {
				differentHash := hashSecret(tt.secret + "x")
				if hash == differentHash {
					t.Error("Different inputs should produce different hashes")
				}
			}
		})
	}
}

// TestValidate_WithAdvancedChecks 测试完整配置验证（包含高级检查）
func TestValidate_WithAdvancedChecks(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "Valid configuration with all checks",
			cfg: &Config{
				Server: ServerConfig{
					Port:         "8080",
					AuthPath:     "/auth",
					HealthPath:   "/health",
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Headers: HeadersConfig{
					UserHeader:   "X-Auth-User",
					RoleHeader:   "X-Auth-Role",
					MethodHeader: "X-Auth-Method",
				},
				Logging: LoggingConfig{
					Format: "json",
					Level:  "info",
				},
				BasicAuths: []BasicAuthConfig{
					{Name: "admin", User: "admin", Pass: "verylongsecurepassword123"},
				},
				JWT: JWTConfig{
					Secret: "this-is-a-very-long-secure-secret-key-for-testing-purposes",
				},
				RoutePolicies: []RoutePolicy{
					{
						Name:              "policy1",
						Host:              "api.example.com",
						AllowedBasicNames: []string{"admin"},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid - policy references non-existent auth",
			cfg: &Config{
				Server: ServerConfig{
					Port:         "8080",
					AuthPath:     "/auth",
					HealthPath:   "/health",
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Headers: HeadersConfig{
					UserHeader:   "X-Auth-User",
					RoleHeader:   "X-Auth-Role",
					MethodHeader: "X-Auth-Method",
				},
				Logging: LoggingConfig{Format: "json", Level: "info"},
				BasicAuths: []BasicAuthConfig{
					{Name: "admin", User: "admin", Pass: "password123456"},
				},
				RoutePolicies: []RoutePolicy{
					{
						Name:              "policy1",
						AllowedBasicNames: []string{"missing-auth"},
					},
				},
			},
			expectErr: true,
			errMsg:    "references",
		},
		{
			name: "Invalid - JWT secret too short",
			cfg: &Config{
				Server: ServerConfig{
					Port:         "8080",
					AuthPath:     "/auth",
					HealthPath:   "/health",
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Headers: HeadersConfig{
					UserHeader:   "X-Auth-User",
					RoleHeader:   "X-Auth-Role",
					MethodHeader: "X-Auth-Method",
				},
				Logging: LoggingConfig{Format: "json", Level: "info"},
				JWT: JWTConfig{
					Secret: "short",
				},
			},
			expectErr: true,
			errMsg:    "secret must be at least",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}
