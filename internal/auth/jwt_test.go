package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// 辅助函数：生成测试 JWT token
func generateTestJWT(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestIsJWT(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "有效的 JWT 格式 (3 部分)",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			want:  true,
		},
		{
			name:  "无效格式 (2 部分)",
			token: "header.payload",
			want:  false,
		},
		{
			name:  "无效格式 (4 部分)",
			token: "part1.part2.part3.part4",
			want:  false,
		},
		{
			name:  "空字符串",
			token: "",
			want:  false,
		},
		{
			name:  "普通字符串",
			token: "not_a_jwt_token",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJWT(tt.token); got != tt.want {
				t.Errorf("IsJWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTryJWT(t *testing.T) {
	secret := "test_secret_key_at_least_32_chars_long_12345"

	tests := []struct {
		name        string
		token       string
		cfg         *config.JWTConfig
		wantSuccess bool
		wantUser    string
		wantRoles   []string
	}{
		{
			name: "有效的 JWT (无额外验证)",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"roles": []interface{}{"user", "admin"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: true,
			wantUser:    "user@example.com",
			wantRoles:   []string{"user", "admin"},
		},
		{
			name: "有效的 JWT (验证 issuer)",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"iss":   "https://auth.example.com",
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
				Issuer: "https://auth.example.com",
			},
			wantSuccess: true,
			wantUser:    "user@example.com",
			wantRoles:   []string{"user"},
		},
		{
			name: "有效的 JWT (验证 audience)",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"aud":   []interface{}{"https://api.example.com"},
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret:   secret,
				Audience: "https://api.example.com",
			},
			wantSuccess: true,
			wantUser:    "user@example.com",
			wantRoles:   []string{"user"},
		},
		{
			name: "JWT issuer 不匹配",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"iss":   "https://wrong.example.com",
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
				Issuer: "https://auth.example.com",
			},
			wantSuccess: false,
		},
		{
			name: "JWT audience 不匹配",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"aud":   []interface{}{"https://wrong.example.com"},
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret:   secret,
				Audience: "https://api.example.com",
			},
			wantSuccess: false,
		},
		{
			name: "JWT 已过期",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(-time.Hour).Unix(), // 过期
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: false,
		},
		{
			name: "错误的签名密钥",
			token: generateTestJWT("wrong_secret_key_12345678901234567890", jwt.MapClaims{
				"sub":   "user@example.com",
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: false,
		},
		{
			name:  "无效的 JWT 格式",
			token: "invalid.jwt.token",
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: false,
		},
		{
			name: "JWT 没有 sub claim",
			token: generateTestJWT(secret, jwt.MapClaims{
				"roles": []interface{}{"user"},
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: false,
		},
		{
			name: "roles 是字符串而不是数组",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub":   "user@example.com",
				"roles": "user,admin", // 字符串格式
				"exp":   time.Now().Add(time.Hour).Unix(),
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: true,
			wantUser:    "user@example.com",
			wantRoles:   []string{}, // 应该优雅处理，返回空角色
		},
		{
			name: "JWT 没有 roles claim (应该返回空角色列表)",
			token: generateTestJWT(secret, jwt.MapClaims{
				"sub": "user@example.com",
				"exp": time.Now().Add(time.Hour).Unix(),
				// 没有 roles claim
			}),
			cfg: &config.JWTConfig{
				Secret: secret,
			},
			wantSuccess: true,
			wantUser:    "user@example.com",
			wantRoles:   []string{}, // 空角色列表
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryJWT(tt.token, tt.cfg)

			if tt.wantSuccess {
				if result == nil {
					t.Fatal("expected success but got nil")
				}
				if result.Method != "jwt" {
					t.Errorf("Method = %q, want %q", result.Method, "jwt")
				}
				if result.User != tt.wantUser {
					t.Errorf("User = %q, want %q", result.User, tt.wantUser)
				}
				if len(result.Roles) != len(tt.wantRoles) {
					t.Errorf("Roles = %v, want %v", result.Roles, tt.wantRoles)
				}
				for i, role := range tt.wantRoles {
					if i >= len(result.Roles) || result.Roles[i] != role {
						t.Errorf("Roles[%d] = %q, want %q", i, result.Roles[i], role)
					}
				}
			} else {
				if result != nil {
					t.Errorf("expected nil but got %+v", result)
				}
			}
		})
	}
}

// 测试边界情况
func TestTryJWT_EdgeCases(t *testing.T) {
	secret := "test_secret_key_at_least_32_chars_long_12345"

	t.Run("空 token", func(t *testing.T) {
		result := TryJWT("", &config.JWTConfig{Secret: secret})
		if result != nil {
			t.Error("empty token should return nil")
		}
	})

	t.Run("空配置", func(t *testing.T) {
		token := generateTestJWT(secret, jwt.MapClaims{
			"sub": "user@example.com",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		result := TryJWT(token, &config.JWTConfig{}) // 空 secret
		if result != nil {
			t.Error("empty secret should return nil")
		}
	})

	t.Run("非常长的 roles 列表", func(t *testing.T) {
		roles := make([]interface{}, 100)
		for i := 0; i < 100; i++ {
			roles[i] = "role" + string(rune(i))
		}
		token := generateTestJWT(secret, jwt.MapClaims{
			"sub":   "user@example.com",
			"roles": roles,
			"exp":   time.Now().Add(time.Hour).Unix(),
		})
		result := TryJWT(token, &config.JWTConfig{Secret: secret})
		if result == nil {
			t.Fatal("should succeed with many roles")
		}
		if len(result.Roles) != 100 {
			t.Errorf("got %d roles, want 100", len(result.Roles))
		}
	})
}

func BenchmarkTryJWT_Valid(b *testing.B) {
	secret := "test_secret_key_at_least_32_chars_long_12345"
	token := generateTestJWT(secret, jwt.MapClaims{
		"sub":   "user@example.com",
		"roles": []interface{}{"user", "admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	cfg := &config.JWTConfig{Secret: secret}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TryJWT(token, cfg)
	}
}

func BenchmarkTryJWT_Invalid(b *testing.B) {
	secret := "test_secret_key_at_least_32_chars_long_12345"
	token := "invalid.jwt.token"
	cfg := &config.JWTConfig{Secret: secret}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TryJWT(token, cfg)
	}
}
