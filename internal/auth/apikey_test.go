package auth

import (
	"testing"

	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

func TestTryAPIKeyAuth(t *testing.T) {
	store := &AuthStore{
		APIKeyByKey: map[string]config.APIKeyConfig{
			"ak_12345": {
				Name:  "internal-service",
				Key:   "ak_12345",
				Roles: []string{"internal", "read"},
			},
			"ak_67890": {
				Name:  "external-service",
				Key:   "ak_67890",
				Roles: []string{"external"},
			},
		},
	}

	tests := []struct {
		name        string
		authHeader  string
		wantSuccess bool
		wantName    string
	}{
		{
			name:        "有效的 API Key (Authorization: ApiKey)",
			authHeader:  "ApiKey ak_12345",
			wantSuccess: true,
			wantName:    "internal-service",
		},
		{
			name:        "另一个有效的 API Key",
			authHeader:  "ApiKey ak_67890",
			wantSuccess: true,
			wantName:    "external-service",
		},
		{
			name:        "无效的 API Key",
			authHeader:  "ApiKey invalid_key",
			wantSuccess: false,
		},
		{
			name:        "空 Key",
			authHeader:  "ApiKey ",
			wantSuccess: false,
		},
		{
			name:        "不是 ApiKey 格式",
			authHeader:  "Bearer token",
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryAPIKeyAuth(tt.authHeader, store)

			if tt.wantSuccess {
				if result == nil {
					t.Fatal("expected success but got nil")
				}
				if result.Method != "apikey" {
					t.Errorf("Method = %q, want %q", result.Method, "apikey")
				}
				if result.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil but got %+v", result)
				}
			}
		})
	}
}

func TestTryAPIKeyHeader(t *testing.T) {
	store := &AuthStore{
		APIKeyByKey: map[string]config.APIKeyConfig{
			"ak_header_123": {
				Name:  "mobile-app",
				Key:   "ak_header_123",
				Roles: []string{"mobile"},
			},
		},
	}

	tests := []struct {
		name        string
		headerValue string
		wantSuccess bool
		wantName    string
	}{
		{
			name:        "有效的 X-Api-Key",
			headerValue: "ak_header_123",
			wantSuccess: true,
			wantName:    "mobile-app",
		},
		{
			name:        "无效的 Key",
			headerValue: "invalid",
			wantSuccess: false,
		},
		{
			name:        "空值",
			headerValue: "",
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryAPIKeyHeader(tt.headerValue, store)

			if tt.wantSuccess {
				if result == nil {
					t.Fatal("expected success but got nil")
				}
				if result.Method != "apikey" {
					t.Errorf("Method = %q, want %q", result.Method, "apikey")
				}
				if result.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil but got %+v", result)
				}
			}
		})
	}
}

func BenchmarkTryAPIKeyAuth(b *testing.B) {
	store := &AuthStore{
		APIKeyByKey: map[string]config.APIKeyConfig{
			"ak_test": {Name: "test", Key: "ak_test", Roles: []string{"test"}},
		},
	}
	authHeader := "ApiKey ak_test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TryAPIKeyAuth(authHeader, store)
	}
}
