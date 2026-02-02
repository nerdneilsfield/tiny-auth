package server

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/policy"
)

// HandleAuth 处理 ForwardAuth 请求
func (s *Server) HandleAuth(c *fiber.Ctx) error {
	// 获取当前配置和存储（线程安全）
	cfg := s.GetConfig()
	store := s.GetStore()

	// 1. 提取 Traefik 转发的原始请求信息
	originalHost := c.Get("X-Forwarded-Host")
	if originalHost == "" {
		originalHost = c.Get("X-Forwarded-Server")
	}
	originalURI := c.Get("X-Forwarded-Uri")
	originalMethod := c.Get("X-Forwarded-Method")
	originalFor := c.Get("X-Forwarded-For")
	if originalFor == "" {
		originalFor = c.IP()
	}

	// 记录请求信息
	fmt.Printf("[Auth] %s %s%s (from %s)\n", originalMethod, originalHost, originalURI, originalFor)

	// 2. 匹配路由策略
	matchedPolicy := policy.MatchPolicy(cfg.RoutePolicies, originalHost, originalURI, originalMethod)

	// 3. 检查是否允许匿名访问
	if matchedPolicy != nil && matchedPolicy.AllowAnonymous {
		return SuccessResponse(c, cfg, &auth.AuthResult{
			Method: "anonymous",
			Roles:  []string{"anonymous"},
		}, matchedPolicy)
	}

	// 4. 尝试各种认证方式（按优先级）
	authHeader := c.Get("Authorization")
	var result *auth.AuthResult

	// 优先级 1: JWT（如果配置了且看起来像 JWT）
	if cfg.JWT.Secret != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if auth.IsJWT(token) {
			result = auth.TryJWT(token, &cfg.JWT)
		}
	}

	// 优先级 2: Bearer Token（静态 token）
	if result == nil && strings.HasPrefix(authHeader, "Bearer ") {
		result = auth.TryBearer(authHeader, store)
	}

	// 优先级 3: Basic Auth
	if result == nil && strings.HasPrefix(authHeader, "Basic ") {
		result = auth.TryBasic(authHeader, store)
	}

	// 优先级 4: API Key (Authorization: ApiKey xxx)
	if result == nil && strings.HasPrefix(authHeader, "ApiKey ") {
		result = auth.TryAPIKeyAuth(authHeader, store)
	}

	// 优先级 5: API Key (X-Api-Key header)
	if result == nil {
		apiKeyHeader := c.Get("X-Api-Key")
		if apiKeyHeader != "" {
			result = auth.TryAPIKeyHeader(apiKeyHeader, store)
		}
	}

	// 5. 检查策略约束
	if result != nil {
		if policy.CheckPolicy(matchedPolicy, result, store) {
			return SuccessResponse(c, cfg, result, matchedPolicy)
		} else {
			// 认证成功但不满足策略要求
			fmt.Printf("[Auth] DENY - policy check failed: %s (method=%s, roles=%v)\n",
				result.User, result.Method, result.Roles)
			return UnauthorizedResponse(c, cfg, "Policy requirements not met")
		}
	}

	// 6. 认证失败
	fmt.Printf("[Auth] DENY - no valid authentication\n")
	return UnauthorizedResponse(c, cfg, "Unauthorized")
}
