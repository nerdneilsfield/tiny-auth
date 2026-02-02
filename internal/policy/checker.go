package policy

import (
	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// CheckPolicy 检查认证结果是否满足策略要求
func CheckPolicy(policy *config.RoutePolicy, result *auth.AuthResult, store *auth.AuthStore) bool {
	if policy == nil {
		return true // 无策略，接受任何有效认证
	}

	// 检查认证方法白名单
	if !checkMethodRestrictions(policy, result) {
		return false
	}

	// 检查角色要求
	if !checkRoleRequirements(policy, result) {
		return false
	}

	return true
}

// checkMethodRestrictions 检查认证方法限制
func checkMethodRestrictions(policy *config.RoutePolicy, result *auth.AuthResult) bool {
	// 🔒 CRITICAL: 全局 JWT Only 检查（必须在其他检查之前）
	// 如果策略要求 JWT Only，非 JWT 请求直接拒绝
	if policy.JWTOnly && result.Method != "jwt" {
		return false
	}

	// 检查具体认证方法的白名单限制
	switch result.Method {
	case "basic":
		// 如果指定了允许的 Basic Auth 名称，检查是否在列表中
		if len(policy.AllowedBasicNames) > 0 {
			return contains(policy.AllowedBasicNames, result.Name)
		}

	case "bearer":
		// 如果指定了允许的 Bearer Token 名称，检查是否在列表中
		if len(policy.AllowedBearerNames) > 0 {
			return contains(policy.AllowedBearerNames, result.Name)
		}

	case "apikey":
		// 如果指定了允许的 API Key 名称，检查是否在列表中
		if len(policy.AllowedAPIKeyNames) > 0 {
			return contains(policy.AllowedAPIKeyNames, result.Name)
		}

	case "jwt":
		// JWT 认证总是允许通过方法检查
		return true
	}

	// 默认：如果没有配置白名单限制，允许通过
	return true
}

// checkRoleRequirements 检查角色要求
func checkRoleRequirements(policy *config.RoutePolicy, result *auth.AuthResult) bool {
	// 检查 require_all_roles：必须拥有所有指定角色
	if len(policy.RequireAllRoles) > 0 {
		for _, required := range policy.RequireAllRoles {
			if !contains(result.Roles, required) {
				return false // 缺少必需角色
			}
		}
	}

	// 检查 require_any_role：必须拥有至少一个指定角色
	if len(policy.RequireAnyRole) > 0 {
		hasAny := false
		for _, required := range policy.RequireAnyRole {
			if contains(result.Roles, required) {
				hasAny = true
				break
			}
		}
		if !hasAny {
			return false // 没有任何必需角色
		}
	}

	return true
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
