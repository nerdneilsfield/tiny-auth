package auth

import (
	"testing"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

func TestTryBearer(t *testing.T) {
	store := &AuthStore{
		BearerByToken: map[string]config.BearerConfig{
			"token_abc123": {
				Name:  "api-service",
				Token: "token_abc123",
				Roles: []string{"api", "read"},
			},
			"token_xyz789": {
				Name:  "admin-service",
				Token: "token_xyz789",
				Roles: []string{"admin", "write"},
			},
		},
	}

	tests := []struct {
		name        string
		authHeader  string
		wantSuccess bool
		wantName    string
		wantRoles   []string
	}{
		{
			name:        "有效的 Bearer Token",
			authHeader:  "Bearer token_abc123",
			wantSuccess: true,
			wantName:    "api-service",
			wantRoles:   []string{"api", "read"},
		},
		{
			name:        "另一个有效的 Bearer Token",
			authHeader:  "Bearer token_xyz789",
			wantSuccess: true,
			wantName:    "admin-service",
			wantRoles:   []string{"admin", "write"},
		},
		{
			name:        "无效的 Token",
			authHeader:  "Bearer invalid_token",
			wantSuccess: false,
		},
		{
			name:        "空 Token",
			authHeader:  "Bearer ",
			wantSuccess: false,
		},
		{
			name:        "不是 Bearer Auth",
			authHeader:  "Basic dGVzdDp0ZXN0",
			wantSuccess: false,
		},
		{
			name:        "空 Header",
			authHeader:  "",
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryBearer(tt.authHeader, store)

			if tt.wantSuccess {
				if result == nil {
					t.Fatal("expected success but got nil")
				}
				if result.Method != "bearer" {
					t.Errorf("Method = %q, want %q", result.Method, "bearer")
				}
				if result.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
				}
				if len(result.Roles) != len(tt.wantRoles) {
					t.Errorf("Roles = %v, want %v", result.Roles, tt.wantRoles)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil but got %+v", result)
				}
			}
		})
	}
}

// 测试常量时间比较
func TestBearerAuth_ConstantTimeComparison(t *testing.T) {
	store := &AuthStore{
		BearerByToken: map[string]config.BearerConfig{
			"secret_token_1234567890abcdef": {
				Name:  "secure-service",
				Token: "secret_token_1234567890abcdef",
				Roles: []string{"secure"},
			},
		},
	}

	wrongTokens := []string{
		"a",
		"wrong",
		"secret_token_1234567890abcdee", // 最后一位不同
		"secret_token_1234567890WRONG",
	}

	for _, token := range wrongTokens {
		result := TryBearer("Bearer "+token, store)
		if result != nil {
			t.Errorf("wrong token %q should fail", token)
		}
	}

	// 正确 token
	result := TryBearer("Bearer secret_token_1234567890abcdef", store)
	if result == nil {
		t.Error("correct token should succeed")
	}
}

func BenchmarkTryBearer(b *testing.B) {
	store := &AuthStore{
		BearerByToken: map[string]config.BearerConfig{
			"token123": {Name: "test", Token: "token123", Roles: []string{"test"}},
		},
	}
	authHeader := "Bearer token123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TryBearer(authHeader, store)
	}
}
