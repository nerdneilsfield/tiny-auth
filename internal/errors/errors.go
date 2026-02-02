package errors

import (
	"errors"
	"fmt"
)

// ErrorCode 错误代码类型
type ErrorCode string

const (
	// Configuration errors
	ErrCodeConfigNotFound      ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigInvalid       ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigValidation    ErrorCode = "CONFIG_VALIDATION"
	ErrCodeConfigPermission    ErrorCode = "CONFIG_PERMISSION"
	ErrCodeEnvVarNotSet        ErrorCode = "ENV_VAR_NOT_SET"
	ErrCodeEnvVarResolution    ErrorCode = "ENV_VAR_RESOLUTION"

	// Authentication errors
	ErrCodeAuthFailed          ErrorCode = "AUTH_FAILED"
	ErrCodeAuthInvalidHeader   ErrorCode = "AUTH_INVALID_HEADER"
	ErrCodeAuthExpired         ErrorCode = "AUTH_EXPIRED"
	ErrCodeAuthInvalidToken    ErrorCode = "AUTH_INVALID_TOKEN"
	ErrCodeAuthInvalidCreds    ErrorCode = "AUTH_INVALID_CREDENTIALS"

	// Authorization errors
	ErrCodeAuthzInsufficientRoles ErrorCode = "AUTHZ_INSUFFICIENT_ROLES"
	ErrCodeAuthzMethodNotAllowed  ErrorCode = "AUTHZ_METHOD_NOT_ALLOWED"
	ErrCodeAuthzJWTRequired       ErrorCode = "AUTHZ_JWT_REQUIRED"

	// Rate limiting errors
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Server errors
	ErrCodeServerStartup ErrorCode = "SERVER_STARTUP"
	ErrCodeServerInternal ErrorCode = "SERVER_INTERNAL"
)

// AppError 应用级错误，包含结构化信息
type AppError struct {
	Code    ErrorCode              // 错误代码
	Message string                 // 人类可读的错误消息
	Details map[string]interface{} // 额外的上下文信息
	Err     error                  // 底层错误（可选）
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is 实现 errors.Is 接口
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithDetail 添加详细信息
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

// Configuration error constructors

func ConfigNotFound(path string) *AppError {
	return NewAppError(
		ErrCodeConfigNotFound,
		"Configuration file not found",
		nil,
	).WithDetail("path", path)
}

func ConfigInvalid(err error) *AppError {
	return NewAppError(
		ErrCodeConfigInvalid,
		"Configuration file is invalid",
		err,
	)
}

func ConfigValidationError(field string, reason string) *AppError {
	return NewAppError(
		ErrCodeConfigValidation,
		"Configuration validation failed",
		nil,
	).WithDetail("field", field).WithDetail("reason", reason)
}

func ConfigPermissionError(path string, mode string) *AppError {
	return NewAppError(
		ErrCodeConfigPermission,
		"Configuration file has insecure permissions",
		nil,
	).WithDetail("path", path).WithDetail("mode", mode)
}

func EnvVarNotSet(varName string) *AppError {
	return NewAppError(
		ErrCodeEnvVarNotSet,
		"Environment variable not set",
		nil,
	).WithDetail("variable", varName)
}

// Authentication error constructors

func AuthFailed(reason string) *AppError {
	return NewAppError(
		ErrCodeAuthFailed,
		"Authentication failed",
		nil,
	).WithDetail("reason", reason)
}

func AuthInvalidHeader(headerName string) *AppError {
	return NewAppError(
		ErrCodeAuthInvalidHeader,
		"Invalid authentication header",
		nil,
	).WithDetail("header", headerName)
}

func AuthExpired(expiresAt string) *AppError {
	return NewAppError(
		ErrCodeAuthExpired,
		"Authentication token expired",
		nil,
	).WithDetail("expires_at", expiresAt)
}

func AuthInvalidToken(reason string) *AppError {
	return NewAppError(
		ErrCodeAuthInvalidToken,
		"Invalid authentication token",
		nil,
	).WithDetail("reason", reason)
}

// Authorization error constructors

func AuthzInsufficientRoles(required []string, actual []string) *AppError {
	return NewAppError(
		ErrCodeAuthzInsufficientRoles,
		"Insufficient roles for access",
		nil,
	).WithDetail("required", required).WithDetail("actual", actual)
}

func AuthzMethodNotAllowed(method string, allowed []string) *AppError {
	return NewAppError(
		ErrCodeAuthzMethodNotAllowed,
		"Authentication method not allowed",
		nil,
	).WithDetail("method", method).WithDetail("allowed", allowed)
}

func AuthzJWTRequired() *AppError {
	return NewAppError(
		ErrCodeAuthzJWTRequired,
		"JWT authentication required",
		nil,
	)
}

// Rate limiting error constructors

func RateLimitExceeded(retryAfter string) *AppError {
	return NewAppError(
		ErrCodeRateLimitExceeded,
		"Rate limit exceeded",
		nil,
	).WithDetail("retry_after", retryAfter)
}

// Server error constructors

func ServerStartupError(err error) *AppError {
	return NewAppError(
		ErrCodeServerStartup,
		"Server startup failed",
		err,
	)
}

func ServerInternalError(err error) *AppError {
	return NewAppError(
		ErrCodeServerInternal,
		"Internal server error",
		err,
	)
}

// Helper functions

// IsConfigError 检查是否为配置错误
func IsConfigError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case ErrCodeConfigNotFound, ErrCodeConfigInvalid, ErrCodeConfigValidation,
			ErrCodeConfigPermission, ErrCodeEnvVarNotSet, ErrCodeEnvVarResolution:
			return true
		}
	}
	return false
}

// IsAuthError 检查是否为认证错误
func IsAuthError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case ErrCodeAuthFailed, ErrCodeAuthInvalidHeader, ErrCodeAuthExpired,
			ErrCodeAuthInvalidToken, ErrCodeAuthInvalidCreds:
			return true
		}
	}
	return false
}

// IsAuthzError 检查是否为授权错误
func IsAuthzError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case ErrCodeAuthzInsufficientRoles, ErrCodeAuthzMethodNotAllowed, ErrCodeAuthzJWTRequired:
			return true
		}
	}
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

// GetErrorDetails 获取错误详情
func GetErrorDetails(err error) map[string]interface{} {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Details
	}
	return nil
}
