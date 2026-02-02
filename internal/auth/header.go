package auth

import "strings"

// ParseAuthHeader 解析 Authorization header，返回 scheme 与 token
func ParseAuthHeader(authHeader string) (scheme, token string) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	scheme = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		token = strings.TrimSpace(parts[1])
	}

	return scheme, token
}
