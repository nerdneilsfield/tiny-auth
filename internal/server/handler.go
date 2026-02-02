package server

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/policy"
	"go.uber.org/zap"
)

// HandleAuth 处理 ForwardAuth 请求
func (s *Server) HandleAuth(c *fiber.Ctx) error {
	startTime := time.Now()

	// 获取当前配置和存储（线程安全）
	cfg := s.GetConfig()
	store := s.GetStore()

	// 1. 安全地提取请求信息（验证可信代理）
	s.mu.RLock()
	trustedCIDRs := s.trustedCIDRs
	rateLimiter := s.RateLimiter
	s.mu.RUnlock()

	// 获取真实客户端 IP（只信任来自可信代理的 X-Forwarded-For）
	clientIP := getClientIP(c, cfg, trustedCIDRs)

	// 2. 速率限制检查
	if rateLimiter != nil {
		allowed, retryAfter := rateLimiter.Allow(clientIP)
		if !allowed {
			s.Logger.Warn("rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.Duration("retry_after", retryAfter),
			)
			c.Set("Retry-After", retryAfter.String())
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many authentication attempts",
				"retry_after": retryAfter.Seconds(),
				"timestamp":   time.Now().Unix(),
			})
		}
	}

	// 获取转发的 headers（只信任来自可信代理的 X-Forwarded-*）
	originalHost, originalURI, originalMethod, trusted := getForwardedHeaders(c, trustedCIDRs)
	requestID := c.Get("X-Request-ID")

	// 构建基础日志字段
	logFields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("client_ip", clientIP),
		zap.String("direct_ip", c.IP()),
		zap.Bool("trusted_proxy", trusted),
		zap.String("method", originalMethod),
		zap.String("host", originalHost),
		zap.String("uri", originalURI),
	}

	// 如果不是来自可信代理，记录警告
	if !trusted && len(trustedCIDRs) > 0 {
		s.Logger.Warn("untrusted proxy detected - using direct connection info",
			append(logFields, zap.String("warning", "X-Forwarded-* headers ignored"))...,
		)
	}

	// 2. 匹配路由策略
	matchedPolicy := policy.MatchPolicy(cfg.RoutePolicies, originalHost, originalURI, originalMethod)

	// 3. 检查是否允许匿名访问
	if matchedPolicy != nil && matchedPolicy.AllowAnonymous {
		s.Logger.Info("auth success - anonymous",
			append(logFields,
				zap.String("auth_method", "anonymous"),
				zap.String("policy", matchedPolicy.Name),
				zap.Duration("latency", time.Since(startTime)),
			)...,
		)
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
			// 认证成功，重置速率限制
			if rateLimiter != nil {
				rateLimiter.Reset(clientIP)
			}

			policyName := ""
			if matchedPolicy != nil {
				policyName = matchedPolicy.Name
			}
			s.Logger.Info("auth success",
				append(logFields,
					zap.String("auth_method", result.Method),
					zap.String("user", result.User),
					zap.Strings("roles", result.Roles),
					zap.String("policy", policyName),
					zap.Duration("latency", time.Since(startTime)),
				)...,
			)
			return SuccessResponse(c, cfg, result, matchedPolicy)
		} else {
			// 认证成功但不满足策略要求
			policyName := ""
			if matchedPolicy != nil {
				policyName = matchedPolicy.Name
			}
			s.Logger.Warn("auth denied - policy check failed",
				append(logFields,
					zap.String("auth_method", result.Method),
					zap.String("user", result.User),
					zap.Strings("roles", result.Roles),
					zap.String("policy", policyName),
					zap.String("reason", "policy_requirements_not_met"),
					zap.Duration("latency", time.Since(startTime)),
				)...,
			)
			return UnauthorizedResponse(c, cfg, "Policy requirements not met")
		}
	}

	// 6. 认证失败
	s.Logger.Warn("auth denied - no valid authentication",
		append(logFields,
			zap.String("reason", "invalid_credentials"),
			zap.Duration("latency", time.Since(startTime)),
		)...,
	)
	return UnauthorizedResponse(c, cfg, "Unauthorized")
}
