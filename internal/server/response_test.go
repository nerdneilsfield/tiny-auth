package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// TestSuccessResponse 测试成功响应
func TestSuccessResponse(t *testing.T) {
	app := fiber.New()

	cfg := &config.Config{
		Headers: config.HeadersConfig{
			MethodHeader: "X-Auth-Method",
			UserHeader:   "X-Auth-User",
			RoleHeader:   "X-Auth-Roles",
		},
	}

	result := &auth.AuthResult{
		Method: "basic",
		User:   "testuser",
		Name:   "Test User",
		Roles:  []string{"admin", "user"},
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessResponse(c, cfg, result, nil)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// 验证状态码
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 验证 headers
	if method := resp.Header.Get("X-Auth-Method"); method != "basic" {
		t.Errorf("Expected X-Auth-Method=basic, got %s", method)
	}

	if user := resp.Header.Get("X-Auth-User"); user != "testuser" {
		t.Errorf("Expected X-Auth-User=testuser, got %s", user)
	}

	if roles := resp.Header.Get("X-Auth-Roles"); roles != "admin,user" {
		t.Errorf("Expected X-Auth-Roles=admin,user, got %s", roles)
	}

	// 验证响应体
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("Expected body 'ok', got %s", string(body))
	}
}

// TestSuccessResponse_WithPolicy 测试带策略的成功响应
func TestSuccessResponse_WithPolicy(t *testing.T) {
	app := fiber.New()

	cfg := &config.Config{
		Headers: config.HeadersConfig{
			MethodHeader: "X-Auth-Method",
		},
	}

	result := &auth.AuthResult{
		Method: "basic",
		User:   "testuser",
	}

	policy := &config.RoutePolicy{
		Name:                "test-policy",
		InjectAuthorization: "Bearer injected-token",
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessResponse(c, cfg, result, policy)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// 验证注入的 Authorization header
	if auth := resp.Header.Get("Authorization"); auth != "Bearer injected-token" {
		t.Errorf("Expected Authorization='Bearer injected-token', got %s", auth)
	}
}

// TestSuccessResponse_WithMetadata 测试带 JWT metadata 的成功响应
func TestSuccessResponse_WithMetadata(t *testing.T) {
	app := fiber.New()

	cfg := &config.Config{
		Headers: config.HeadersConfig{
			MethodHeader:       "X-Auth-Method",
			IncludeJWTMetadata: true,
		},
	}

	result := &auth.AuthResult{
		Method: "jwt",
		User:   "testuser",
		Metadata: map[string]string{
			"issuer":   "auth-service",
			"audience": "api-service",
		},
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessResponse(c, cfg, result, nil)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// 验证 metadata headers
	if issuer := resp.Header.Get("X-Auth-Issuer"); issuer != "auth-service" {
		t.Errorf("Expected X-Auth-Issuer=auth-service, got %s", issuer)
	}

	if audience := resp.Header.Get("X-Auth-Audience"); audience != "api-service" {
		t.Errorf("Expected X-Auth-Audience=api-service, got %s", audience)
	}
}

// TestSuccessResponse_ExtraHeaders 测试额外的 headers
func TestSuccessResponse_ExtraHeaders(t *testing.T) {
	app := fiber.New()

	cfg := &config.Config{
		Headers: config.HeadersConfig{
			MethodHeader: "X-Auth-Method",
			ExtraHeaders: []string{"X-Auth-Timestamp", "X-Auth-Route"},
		},
	}

	result := &auth.AuthResult{
		Method: "basic",
		User:   "testuser",
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		// 模拟 X-Forwarded-* headers
		c.Request().Header.Set("X-Forwarded-Host", "api.example.com")
		c.Request().Header.Set("X-Forwarded-Uri", "/api/data")
		return SuccessResponse(c, cfg, result, nil)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// 验证额外的 headers
	if timestamp := resp.Header.Get("X-Auth-Timestamp"); timestamp == "" {
		t.Error("Expected X-Auth-Timestamp to be set")
	}

	if route := resp.Header.Get("X-Auth-Route"); route != "api.example.com/api/data" {
		t.Errorf("Expected X-Auth-Route=api.example.com/api/data, got %s", route)
	}
}

// TestUnauthorizedResponse 测试未授权响应
func TestUnauthorizedResponse(t *testing.T) {
	app := fiber.New()

	cfg := &config.Config{
		BasicAuths: []config.BasicAuthConfig{
			{Name: "test", User: "test", Pass: "test"},
		},
		BearerTokens: []config.BearerConfig{
			{Name: "test", Token: "test-token"},
		},
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return UnauthorizedResponse(c, cfg, "Invalid credentials")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// 验证状态码
	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// 验证 WWW-Authenticate headers
	// Note: Fiber's Append may join multiple values with comma or create multiple headers
	// depending on the HTTP/1.1 spec and implementation
	authHeader := resp.Header.Get("Www-Authenticate")
	authHeaders := resp.Header["Www-Authenticate"]

	// 合并所有 WWW-Authenticate headers
	allAuthMethods := strings.Join(authHeaders, ", ")

	// 验证至少包含 Basic 和 Bearer
	if !strings.Contains(allAuthMethods, "Basic") {
		t.Errorf("Expected WWW-Authenticate to include Basic, got: %s", allAuthMethods)
	}
	if !strings.Contains(allAuthMethods, "Bearer") {
		t.Errorf("Expected WWW-Authenticate to include Bearer, got: %s", allAuthMethods)
	}

	// Debug output (optional)
	if len(authHeaders) > 0 {
		t.Logf("WWW-Authenticate headers: %v (combined: %s)", authHeaders, allAuthMethods)
	} else if authHeader != "" {
		t.Logf("WWW-Authenticate header (single): %s", authHeader)
	}

	// 验证 Cache-Control
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Expected Cache-Control=no-store, got %s", cc)
	}

	// 验证响应体
	body, _ := io.ReadAll(resp.Body)
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if errMsg := jsonResp["error"]; errMsg != "Invalid credentials" {
		t.Errorf("Expected error='Invalid credentials', got %v", errMsg)
	}

	if _, ok := jsonResp["timestamp"]; !ok {
		t.Error("Expected timestamp in response")
	}
}

// TestSanitizeHeaderValue 测试 header 值清理
func TestSanitizeHeaderValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Normal value",
			input: "normal-value",
			want:  "normal-value",
		},
		{
			name:  "With newline",
			input: "value\nwith\nnewline",
			want:  "valuewithnewline",
		},
		{
			name:  "With carriage return",
			input: "value\rwith\rcarriage",
			want:  "valuewithcarriage",
		},
		{
			name:  "With CRLF",
			input: "value\r\nwith\r\ncrlf",
			want:  "valuewithcrlf",
		},
		{
			name:  "Very long value",
			input: strings.Repeat("a", 2000),
			want:  strings.Repeat("a", 1024),
		},
		{
			name:  "Header injection attempt",
			input: "value\r\nX-Injected-Header: malicious",
			want:  "valueX-Injected-Header: malicious",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeHeaderValue(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeHeaderValue() = %q, want %q", got, tt.want)
			}

			// 验证清理后的值不包含换行符
			if strings.Contains(got, "\n") || strings.Contains(got, "\r") {
				t.Error("Sanitized value still contains newline characters")
			}

			// 验证清理后的值不超过 1024 字节
			if len(got) > 1024 {
				t.Errorf("Sanitized value length %d exceeds 1024", len(got))
			}
		})
	}
}

// TestSanitizeHeaderValue_EmptyString 测试空字符串
func TestSanitizeHeaderValue_EmptyString(t *testing.T) {
	result := sanitizeHeaderValue("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// TestSanitizeHeaderValue_ExactlyMaxLength 测试恰好最大长度
func TestSanitizeHeaderValue_ExactlyMaxLength(t *testing.T) {
	input := strings.Repeat("a", 1024)
	result := sanitizeHeaderValue(input)
	if len(result) != 1024 {
		t.Errorf("Expected length 1024, got %d", len(result))
	}
	if result != input {
		t.Error("Value should not be truncated when exactly at max length")
	}
}
