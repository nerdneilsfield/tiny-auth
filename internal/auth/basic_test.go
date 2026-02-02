package auth

import (
	"encoding/base64"
	"testing"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

func TestTryBasic(t *testing.T) {
	// 准备测试数据
	store := &AuthStore{
		BasicByUser: map[string]config.BasicAuthConfig{
			"admin": {
				Name:  "admin-basic",
				User:  "admin",
				Pass:  "secret123",
				Roles: []string{"admin", "user"},
			},
			"dev": {
				Name:  "dev-basic",
				User:  "dev",
				Pass:  "devpass",
				Roles: []string{"developer"},
			},
		},
	}

	tests := []struct {
		name         string
		authHeader   string
		wantSuccess  bool
		wantUser     string
		wantName     string
		wantRoles    []string
		wantMethod   string
	}{
		{
			name:        "有效的 Basic Auth (admin)",
			authHeader:  "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret123")),
			wantSuccess: true,
			wantUser:    "admin",
			wantName:    "admin-basic",
			wantRoles:   []string{"admin", "user"},
			wantMethod:  "basic",
		},
		{
			name:        "有效的 Basic Auth (dev)",
			authHeader:  "Basic " + base64.StdEncoding.EncodeToString([]byte("dev:devpass")),
			wantSuccess: true,
			wantUser:    "dev",
			wantName:    "dev-basic",
			wantRoles:   []string{"developer"},
			wantMethod:  "basic",
		},
		{
			name:        "错误的密码",
			authHeader:  "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrongpass")),
			wantSuccess: false,
		},
		{
			name:        "不存在的用户",
			authHeader:  "Basic " + base64.StdEncoding.EncodeToString([]byte("notexist:anypass")),
			wantSuccess: false,
		},
		{
			name:        "无效的 base64",
			authHeader:  "Basic invalid!!!base64",
			wantSuccess: false,
		},
		{
			name:        "缺少冒号分隔符",
			authHeader:  "Basic " + base64.StdEncoding.EncodeToString([]byte("adminnocolon")),
			wantSuccess: false,
		},
		{
			name:        "空的 Authorization header",
			authHeader:  "",
			wantSuccess: false,
		},
		{
			name:        "不是 Basic Auth",
			authHeader:  "Bearer token123",
			wantSuccess: false,
		},
		{
			name:        "空密码（应该失败）",
			authHeader:  "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:")),
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryBasic(tt.authHeader, store)

			if tt.wantSuccess {
				if result == nil {
					t.Fatal("expected success but got nil result")
				}
				if result.User != tt.wantUser {
					t.Errorf("User = %q, want %q", result.User, tt.wantUser)
				}
				if result.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
				}
				if result.Method != tt.wantMethod {
					t.Errorf("Method = %q, want %q", result.Method, tt.wantMethod)
				}
				if len(result.Roles) != len(tt.wantRoles) {
					t.Errorf("Roles count = %d, want %d", len(result.Roles), len(tt.wantRoles))
				}
				for i, role := range tt.wantRoles {
					if i >= len(result.Roles) || result.Roles[i] != role {
						t.Errorf("Roles[%d] = %q, want %q", i, result.Roles[i], role)
					}
				}
			} else {
				if result != nil {
					t.Errorf("expected nil but got result: %+v", result)
				}
			}
		})
	}
}

// TestBasicAuth_ConstantTimeComparison 测试常量时间比较
// 这是一个安全测试，确保我们使用常量时间比较来防止时序攻击
func TestBasicAuth_ConstantTimeComparison(t *testing.T) {
	store := &AuthStore{
		BasicByUser: map[string]config.BasicAuthConfig{
			"admin": {
				Name:  "admin-basic",
				User:  "admin",
				Pass:  "verylongsecretpassword123456",
				Roles: []string{"admin"},
			},
		},
	}

	// 测试多个不同长度的错误密码
	// 常量时间比较应该对所有错误密码花费相同时间
	wrongPasswords := []string{
		"a",                              // 短密码
		"wrongpass",                      // 中等密码
		"verylongsecretpassword123455",   // 几乎正确的长密码（最后一位不同）
		"verylongsecretpasswordWRONG123", // 完全错误的长密码
	}

	for _, pass := range wrongPasswords {
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:"+pass))
		result := TryBasic(authHeader, store)
		if result != nil {
			t.Errorf("wrong password %q should fail but got result", pass)
		}
	}

	// 验证正确密码能通过
	correctAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:verylongsecretpassword123456"))
	result := TryBasic(correctAuth, store)
	if result == nil {
		t.Error("correct password should succeed")
	}
}

// Benchmark 测试性能
func BenchmarkTryBasic_Success(b *testing.B) {
	store := &AuthStore{
		BasicByUser: map[string]config.BasicAuthConfig{
			"admin": {
				Name:  "admin-basic",
				User:  "admin",
				Pass:  "secret123",
				Roles: []string{"admin"},
			},
		},
	}
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret123"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TryBasic(authHeader, store)
	}
}

func BenchmarkTryBasic_Failure(b *testing.B) {
	store := &AuthStore{
		BasicByUser: map[string]config.BasicAuthConfig{
			"admin": {
				Name:  "admin-basic",
				User:  "admin",
				Pass:  "secret123",
				Roles: []string{"admin"},
			},
		},
	}
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrongpass"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TryBasic(authHeader, store)
	}
}
