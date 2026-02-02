package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

// TryBasic 尝试 Basic Auth 认证
func TryBasic(authHeader string, store *AuthStore) *AuthResult {
	if !strings.HasPrefix(authHeader, "Basic ") {
		return nil
	}

	// 解码 base64
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
	if err != nil {
		return nil
	}

	// 解析 username:password
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return nil
	}

	user, pass := parts[0], parts[1]

	// 查找用户配置
	cfg, ok := store.BasicByUser[user]
	if !ok {
		return nil
	}

	// 使用常量时间比较密码（防止时序攻击）
	if subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.Pass)) != 1 {
		return nil
	}

	// 认证成功
	return &AuthResult{
		Method: "basic",
		Name:   cfg.Name,
		User:   user,
		Roles:  cfg.Roles,
	}
}
