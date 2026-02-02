package errors

import (
	"errors"
	"strings"
	"testing"
)

// TestAppError_Error 测试 Error 方法
func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		contains []string
	}{
		{
			name: "Error without underlying error",
			err: &AppError{
				Code:    ErrCodeAuthFailed,
				Message: "Authentication failed",
			},
			contains: []string{"AUTH_FAILED", "Authentication failed"},
		},
		{
			name: "Error with underlying error",
			err: &AppError{
				Code:    ErrCodeConfigInvalid,
				Message: "Invalid configuration",
				Err:     errors.New("parsing error"),
			},
			contains: []string{"CONFIG_INVALID", "Invalid configuration", "parsing error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("Error message %q does not contain %q", errMsg, substr)
				}
			}
		})
	}
}

// TestAppError_Unwrap 测试 Unwrap 方法
func TestAppError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	appErr := &AppError{
		Code:    ErrCodeServerInternal,
		Message: "Server error",
		Err:     underlying,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Expected unwrapped error %v, got %v", underlying, unwrapped)
	}
}

// TestAppError_Is 测试 Is 方法
func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Code: ErrCodeAuthFailed, Message: "Failed"}
	err2 := &AppError{Code: ErrCodeAuthFailed, Message: "Also failed"}
	err3 := &AppError{Code: ErrCodeAuthExpired, Message: "Expired"}

	if !errors.Is(err1, err2) {
		t.Error("Expected err1 to match err2 (same code)")
	}

	if errors.Is(err1, err3) {
		t.Error("Expected err1 not to match err3 (different code)")
	}
}

// TestAppError_WithDetail 测试 WithDetail 方法
func TestAppError_WithDetail(t *testing.T) {
	err := NewAppError(ErrCodeAuthFailed, "Failed", nil)
	
	err.WithDetail("user", "admin").WithDetail("ip", "192.168.1.1")

	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}

	if err.Details["user"] != "admin" {
		t.Errorf("Expected user='admin', got %v", err.Details["user"])
	}

	if err.Details["ip"] != "192.168.1.1" {
		t.Errorf("Expected ip='192.168.1.1', got %v", err.Details["ip"])
	}
}

// TestConfigNotFound 测试 ConfigNotFound 构造函数
func TestConfigNotFound(t *testing.T) {
	err := ConfigNotFound("/path/to/config.toml")

	if err.Code != ErrCodeConfigNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeConfigNotFound, err.Code)
	}

	if err.Details["path"] != "/path/to/config.toml" {
		t.Errorf("Expected path in details, got %v", err.Details)
	}
}

// TestConfigValidationError 测试 ConfigValidationError 构造函数
func TestConfigValidationError(t *testing.T) {
	err := ConfigValidationError("server.port", "must be a valid port number")

	if err.Code != ErrCodeConfigValidation {
		t.Errorf("Expected code %s, got %s", ErrCodeConfigValidation, err.Code)
	}

	if err.Details["field"] != "server.port" {
		t.Errorf("Expected field in details")
	}

	if err.Details["reason"] != "must be a valid port number" {
		t.Errorf("Expected reason in details")
	}
}

// TestAuthzInsufficientRoles 测试 AuthzInsufficientRoles 构造函数
func TestAuthzInsufficientRoles(t *testing.T) {
	required := []string{"admin", "moderator"}
	actual := []string{"user"}

	err := AuthzInsufficientRoles(required, actual)

	if err.Code != ErrCodeAuthzInsufficientRoles {
		t.Errorf("Expected code %s, got %s", ErrCodeAuthzInsufficientRoles, err.Code)
	}

	if err.Details["required"] == nil {
		t.Error("Expected required roles in details")
	}

	if err.Details["actual"] == nil {
		t.Error("Expected actual roles in details")
	}
}

// TestRateLimitExceeded 测试 RateLimitExceeded 构造函数
func TestRateLimitExceeded(t *testing.T) {
	err := RateLimitExceeded("300s")

	if err.Code != ErrCodeRateLimitExceeded {
		t.Errorf("Expected code %s, got %s", ErrCodeRateLimitExceeded, err.Code)
	}

	if err.Details["retry_after"] != "300s" {
		t.Errorf("Expected retry_after in details")
	}
}

// TestIsConfigError 测试 IsConfigError 函数
func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Config not found",
			err:      ConfigNotFound("/path"),
			expected: true,
		},
		{
			name:     "Config validation error",
			err:      ConfigValidationError("field", "reason"),
			expected: true,
		},
		{
			name:     "Auth error",
			err:      AuthFailed("reason"),
			expected: false,
		},
		{
			name:     "Standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConfigError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected IsConfigError=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsAuthError 测试 IsAuthError 函数
func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Auth failed",
			err:      AuthFailed("bad password"),
			expected: true,
		},
		{
			name:     "Auth expired",
			err:      AuthExpired("2024-01-01"),
			expected: true,
		},
		{
			name:     "Config error",
			err:      ConfigNotFound("/path"),
			expected: false,
		},
		{
			name:     "Standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected IsAuthError=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsAuthzError 测试 IsAuthzError 函数
func TestIsAuthzError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Insufficient roles",
			err:      AuthzInsufficientRoles([]string{"admin"}, []string{"user"}),
			expected: true,
		},
		{
			name:     "JWT required",
			err:      AuthzJWTRequired(),
			expected: true,
		},
		{
			name:     "Auth error",
			err:      AuthFailed("reason"),
			expected: false,
		},
		{
			name:     "Standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthzError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected IsAuthzError=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetErrorCode 测试 GetErrorCode 函数
func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "AppError with code",
			err:      AuthFailed("reason"),
			expected: ErrCodeAuthFailed,
		},
		{
			name:     "Standard error",
			err:      errors.New("standard"),
			expected: "",
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCode(tt.err)
			if result != tt.expected {
				t.Errorf("Expected code %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGetErrorDetails 测试 GetErrorDetails 函数
func TestGetErrorDetails(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectNil   bool
		expectKey   string
		expectValue interface{}
	}{
		{
			name:        "AppError with details",
			err:         ConfigNotFound("/path/to/config"),
			expectNil:   false,
			expectKey:   "path",
			expectValue: "/path/to/config",
		},
		{
			name:      "Standard error",
			err:       errors.New("standard"),
			expectNil: true,
		},
		{
			name:      "Nil error",
			err:       nil,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := GetErrorDetails(tt.err)

			if tt.expectNil {
				if details != nil {
					t.Error("Expected nil details")
				}
				return
			}

			if details == nil {
				t.Fatal("Expected non-nil details")
			}

			if tt.expectKey != "" {
				if details[tt.expectKey] != tt.expectValue {
					t.Errorf("Expected details[%q]=%v, got %v",
						tt.expectKey, tt.expectValue, details[tt.expectKey])
				}
			}
		})
	}
}

// TestErrorWrapping 测试错误包装
func TestErrorWrapping(t *testing.T) {
	underlying := errors.New("underlying error")
	wrapped := ConfigInvalid(underlying)

	// 测试 errors.Is
	if !errors.Is(wrapped, wrapped) {
		t.Error("Error should match itself")
	}

	// 测试 errors.As
	var appErr *AppError
	if !errors.As(wrapped, &appErr) {
		t.Error("Should be able to unwrap to AppError")
	}

	if appErr.Code != ErrCodeConfigInvalid {
		t.Errorf("Expected code %s, got %s", ErrCodeConfigInvalid, appErr.Code)
	}

	// 测试 errors.Unwrap
	if !errors.Is(wrapped, underlying) {
		t.Error("Should be able to unwrap to underlying error")
	}
}
