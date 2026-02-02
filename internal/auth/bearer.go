package auth

import (
	"crypto/subtle"
	"strings"
)

// TryBearer 尝试 Bearer Token 认证
func TryBearer(authHeader string, store *AuthStore) *AuthResult {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		return nil
	}

	// 查找 token 配置
	for storedToken, cfg := range store.BearerByToken {
		// 使用常量时间比较（防止时序攻击）
		if subtle.ConstantTimeCompare([]byte(token), []byte(storedToken)) == 1 {
			return &AuthResult{
				Method: "bearer",
				Name:   cfg.Name,
				Roles:  cfg.Roles,
			}
		}
	}

	return nil
}
