package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// TryBasic 尝试 Basic Auth 认证
func TryBasic(authHeader string, store *AuthStore) *AuthResult {
	scheme, payload := ParseAuthHeader(authHeader)
	if !strings.EqualFold(scheme, "Basic") {
		return nil
	}

	// 解码 base64
	if payload == "" {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil
	}

	// 解析 username:password
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil
	}

	user, pass := parts[0], parts[1]

	// 查找用户配置
	cfg, ok := store.BasicByUser[user]
	if !ok {
		return nil
	}

	// 验证密码
	// 优先使用 bcrypt 哈希（如果配置了）
	var passwordValid bool
	if cfg.PassHash != "" {
		// 使用 bcrypt 验证哈希密码
		err := bcrypt.CompareHashAndPassword([]byte(cfg.PassHash), []byte(pass))
		passwordValid = (err == nil)
	} else {
		// 回退到明文密码比较（使用常量时间比较防止时序攻击）
		passwordValid = (subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.Pass)) == 1)
	}

	if !passwordValid {
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
