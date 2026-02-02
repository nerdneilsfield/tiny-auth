package auth

import (
	"crypto/subtle"
	"strings"
)

// TryAPIKeyAuth 尝试 API Key 认证（通过 Authorization: ApiKey xxx）
func TryAPIKeyAuth(authHeader string, store *AuthStore) *AuthResult {
	if !strings.HasPrefix(authHeader, "ApiKey ") {
		return nil
	}

	key := strings.TrimPrefix(authHeader, "ApiKey ")
	key = strings.TrimSpace(key)

	if key == "" {
		return nil
	}

	return lookupAPIKey(key, store)
}

// TryAPIKeyHeader 尝试 API Key 认证（通过 X-Api-Key header）
func TryAPIKeyHeader(headerValue string, store *AuthStore) *AuthResult {
	if headerValue == "" {
		return nil
	}

	return lookupAPIKey(headerValue, store)
}

// lookupAPIKey 在存储中查找 API Key
func lookupAPIKey(key string, store *AuthStore) *AuthResult {
	// 查找 key 配置
	for storedKey, cfg := range store.APIKeyByKey {
		// 使用常量时间比较（防止时序攻击）
		if subtle.ConstantTimeCompare([]byte(key), []byte(storedKey)) == 1 {
			return &AuthResult{
				Method: "apikey",
				Name:   cfg.Name,
				Roles:  cfg.Roles,
			}
		}
	}

	return nil
}
