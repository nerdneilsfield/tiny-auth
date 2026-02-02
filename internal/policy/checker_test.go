package policy

import (
	"testing"

	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// TestCheckMethodRestrictions_JWTOnly 测试 jwt_only 策略绕过漏洞的修复
// 这是一个关键的安全测试用例
func TestCheckMethodRestrictions_JWTOnly(t *testing.T) {
	tests := []struct {
		name     string
		policy   *config.RoutePolicy
		result   *auth.AuthResult
		expected bool
	}{
		{
			name: "jwt_only=true 应该拒绝 Basic Auth",
			policy: &config.RoutePolicy{
				JWTOnly: true,
			},
			result: &auth.AuthResult{
				Method: "basic",
				User:   "admin",
				Name:   "admin-basic",
				Roles:  []string{"admin"},
			},
			expected: false, // 必须拒绝
		},
		{
			name: "jwt_only=true 应该拒绝 Bearer Token",
			policy: &config.RoutePolicy{
				JWTOnly: true,
			},
			result: &auth.AuthResult{
				Method: "bearer",
				Name:   "api-token",
				Roles:  []string{"api"},
			},
			expected: false, // 必须拒绝
		},
		{
			name: "jwt_only=true 应该拒绝 API Key",
			policy: &config.RoutePolicy{
				JWTOnly: true,
			},
			result: &auth.AuthResult{
				Method: "apikey",
				Name:   "internal-key",
				Roles:  []string{"internal"},
			},
			expected: false, // 必须拒绝
		},
		{
			name: "jwt_only=true 应该允许 JWT",
			policy: &config.RoutePolicy{
				JWTOnly: true,
			},
			result: &auth.AuthResult{
				Method: "jwt",
				User:   "user@example.com",
				Roles:  []string{"user"},
			},
			expected: true, // 允许
		},
		{
			name: "jwt_only=false 应该允许 Basic Auth",
			policy: &config.RoutePolicy{
				JWTOnly: false,
			},
			result: &auth.AuthResult{
				Method: "basic",
				User:   "admin",
				Name:   "admin-basic",
				Roles:  []string{"admin"},
			},
			expected: true, // 允许
		},
		{
			name: "白名单限制应该生效（在白名单内）",
			policy: &config.RoutePolicy{
				AllowedBasicNames: []string{"admin-basic", "dev-basic"},
			},
			result: &auth.AuthResult{
				Method: "basic",
				User:   "admin",
				Name:   "admin-basic",
				Roles:  []string{"admin"},
			},
			expected: true, // 允许
		},
		{
			name: "白名单限制应该生效（不在白名单内）",
			policy: &config.RoutePolicy{
				AllowedBasicNames: []string{"admin-basic"},
			},
			result: &auth.AuthResult{
				Method: "basic",
				User:   "dev",
				Name:   "dev-basic",
				Roles:  []string{"dev"},
			},
			expected: false, // 拒绝
		},
		{
			name: "jwt_only + 白名单冲突时，jwt_only 优先",
			policy: &config.RoutePolicy{
				JWTOnly:           true,
				AllowedBasicNames: []string{"admin-basic"}, // 即使在白名单，也应该被 jwt_only 拒绝
			},
			result: &auth.AuthResult{
				Method: "basic",
				User:   "admin",
				Name:   "admin-basic",
				Roles:  []string{"admin"},
			},
			expected: false, // 必须拒绝（jwt_only 优先）
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkMethodRestrictions(tt.policy, tt.result)
			if got != tt.expected {
				t.Errorf("checkMethodRestrictions() = %v, want %v", got, tt.expected)
				t.Errorf("  Policy: jwt_only=%v, allowed_basic=%v",
					tt.policy.JWTOnly, tt.policy.AllowedBasicNames)
				t.Errorf("  Result: method=%s, name=%s",
					tt.result.Method, tt.result.Name)
			}
		})
	}
}

// TestCheckRoleRequirements 测试角色要求检查
func TestCheckRoleRequirements(t *testing.T) {
	tests := []struct {
		name     string
		policy   *config.RoutePolicy
		result   *auth.AuthResult
		expected bool
	}{
		{
			name: "require_all_roles - 拥有所有角色",
			policy: &config.RoutePolicy{
				RequireAllRoles: []string{"admin", "user"},
			},
			result: &auth.AuthResult{
				Roles: []string{"admin", "user", "editor"},
			},
			expected: true,
		},
		{
			name: "require_all_roles - 缺少某个角色",
			policy: &config.RoutePolicy{
				RequireAllRoles: []string{"admin", "superuser"},
			},
			result: &auth.AuthResult{
				Roles: []string{"admin", "user"},
			},
			expected: false,
		},
		{
			name: "require_any_role - 拥有其中一个角色",
			policy: &config.RoutePolicy{
				RequireAnyRole: []string{"admin", "moderator"},
			},
			result: &auth.AuthResult{
				Roles: []string{"user", "moderator"},
			},
			expected: true,
		},
		{
			name: "require_any_role - 没有任何角色",
			policy: &config.RoutePolicy{
				RequireAnyRole: []string{"admin", "moderator"},
			},
			result: &auth.AuthResult{
				Roles: []string{"user", "guest"},
			},
			expected: false,
		},
		{
			name: "同时满足 require_all_roles 和 require_any_role",
			policy: &config.RoutePolicy{
				RequireAllRoles: []string{"user"},
				RequireAnyRole:  []string{"admin", "moderator"},
			},
			result: &auth.AuthResult{
				Roles: []string{"user", "admin"},
			},
			expected: true,
		},
		{
			name: "满足 require_any_role 但不满足 require_all_roles",
			policy: &config.RoutePolicy{
				RequireAllRoles: []string{"user", "premium"},
				RequireAnyRole:  []string{"admin", "moderator"},
			},
			result: &auth.AuthResult{
				Roles: []string{"user", "admin"}, // 缺少 premium
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRoleRequirements(tt.policy, tt.result)
			if got != tt.expected {
				t.Errorf("checkRoleRequirements() = %v, want %v", got, tt.expected)
				t.Errorf("  Policy: require_all=%v, require_any=%v",
					tt.policy.RequireAllRoles, tt.policy.RequireAnyRole)
				t.Errorf("  Result: roles=%v", tt.result.Roles)
			}
		})
	}
}

// TestCheckPolicy 测试完整策略检查
func TestCheckPolicy(t *testing.T) {
	tests := []struct {
		name     string
		policy   *config.RoutePolicy
		result   *auth.AuthResult
		expected bool
	}{
		{
			name:   "无策略应该接受任何有效认证",
			policy: nil,
			result: &auth.AuthResult{
				Method: "basic",
				User:   "user",
			},
			expected: true,
		},
		{
			name: "jwt_only + 角色要求",
			policy: &config.RoutePolicy{
				JWTOnly:         true,
				RequireAllRoles: []string{"admin"},
			},
			result: &auth.AuthResult{
				Method: "jwt",
				User:   "admin@example.com",
				Roles:  []string{"admin", "user"},
			},
			expected: true,
		},
		{
			name: "jwt_only + 角色要求（不满足角色）",
			policy: &config.RoutePolicy{
				JWTOnly:         true,
				RequireAllRoles: []string{"admin"},
			},
			result: &auth.AuthResult{
				Method: "jwt",
				User:   "user@example.com",
				Roles:  []string{"user"},
			},
			expected: false,
		},
		{
			name: "jwt_only + 角色要求（不满足方法）",
			policy: &config.RoutePolicy{
				JWTOnly:         true,
				RequireAllRoles: []string{"admin"},
			},
			result: &auth.AuthResult{
				Method: "basic",
				User:   "admin",
				Roles:  []string{"admin"},
			},
			expected: false, // 即使有 admin 角色，但不是 JWT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPolicy(tt.policy, tt.result, nil)
			if got != tt.expected {
				t.Errorf("CheckPolicy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestContains 测试辅助函数
func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "orange"}

	tests := []struct {
		item     string
		expected bool
	}{
		{"apple", true},
		{"banana", true},
		{"orange", true},
		{"grape", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			got := contains(slice, tt.item)
			if got != tt.expected {
				t.Errorf("contains() = %v, want %v", got, tt.expected)
			}
		})
	}
}
