package auth

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// TryJWT 尝试 JWT 认证
func TryJWT(tokenString string, jwtCfg *config.JWTConfig) *AuthResult {
	if jwtCfg.Secret == "" {
		return nil // JWT 未配置
	}

	// 解析并验证 JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtCfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil
	}

	// 提取 claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	// 验证 issuer（如果配置了）
	if jwtCfg.Issuer != "" {
		if iss, _ := claims["iss"].(string); iss != jwtCfg.Issuer {
			return nil
		}
	}

	// 验证 audience（如果配置了）
	if jwtCfg.Audience != "" {
		audMatched := false
		if aud, ok := claims["aud"].(string); ok {
			// aud 是字符串
			audMatched = (aud == jwtCfg.Audience)
		} else if audArray, ok := claims["aud"].([]interface{}); ok {
			// aud 是数组
			for _, a := range audArray {
				if audStr, ok := a.(string); ok && audStr == jwtCfg.Audience {
					audMatched = true
					break
				}
			}
		}
		if !audMatched {
			return nil
		}
	}

	// 提取用户信息
	user, _ := claims["sub"].(string)
	if user == "" {
		return nil // sub claim 是必需的
	}

	// 提取角色（支持 roles 数组或单个 role 字符串）
	var roles []string
	if rolesInterface, ok := claims["roles"]; ok {
		// 处理 roles 数组
		if rolesArray, ok := rolesInterface.([]interface{}); ok {
			for _, r := range rolesArray {
				if roleStr, ok := r.(string); ok && roleStr != "" {
					roles = append(roles, roleStr)
				}
			}
		}
	} else if roleStr, ok := claims["role"].(string); ok && roleStr != "" {
		// 兼容单个 role 字段
		roles = append(roles, roleStr)
	}

	// 构建元数据
	metadata := make(map[string]string)
	if iss, ok := claims["iss"].(string); ok {
		metadata["issuer"] = iss
	}
	if aud, ok := claims["aud"].(string); ok {
		metadata["audience"] = aud
	}

	return &AuthResult{
		Method:   "jwt",
		User:     user,
		Roles:    roles,
		Metadata: metadata,
	}
}

// IsJWT 检查 Bearer token 是否看起来像 JWT（有3段用.分隔）
func IsJWT(token string) bool {
	return len(strings.Split(token, ".")) == 3
}
