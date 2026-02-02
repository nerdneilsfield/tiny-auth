package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// SuccessResponse 返回认证成功响应
func SuccessResponse(c *fiber.Ctx, cfg *config.Config, result *auth.AuthResult, policy *config.RoutePolicy) error {
	setMethodHeader(c, cfg, result)
	setUserHeader(c, cfg, result)
	setRoleHeader(c, cfg, result)
	setExtraHeaders(c, cfg)
	setJWTMetadataHeaders(c, cfg, result)
	setInjectedAuthorization(c, policy)

	// 返回 200 OK
	c.Status(fiber.StatusOK)
	return c.SendString("ok")
}

func setMethodHeader(c *fiber.Ctx, cfg *config.Config, result *auth.AuthResult) {
	if cfg.Headers.MethodHeader == "" {
		return
	}

	// 虽然 method 是系统生成的，但为了一致性也进行清理
	c.Set(cfg.Headers.MethodHeader, sanitizeHeaderValue(result.Method))
}

func setUserHeader(c *fiber.Ctx, cfg *config.Config, result *auth.AuthResult) {
	if cfg.Headers.UserHeader == "" {
		return
	}

	if result.User != "" {
		c.Set(cfg.Headers.UserHeader, sanitizeHeaderValue(result.User))
		return
	}

	if result.Name != "" {
		// 如果没有用户名，使用配置名称
		c.Set(cfg.Headers.UserHeader, sanitizeHeaderValue(result.Name))
	}
}

func setRoleHeader(c *fiber.Ctx, cfg *config.Config, result *auth.AuthResult) {
	if cfg.Headers.RoleHeader == "" || len(result.Roles) == 0 {
		return
	}

	roles := strings.Join(result.Roles, ",")
	c.Set(cfg.Headers.RoleHeader, sanitizeHeaderValue(roles))
}

func setExtraHeaders(c *fiber.Ctx, cfg *config.Config) {
	for _, h := range cfg.Headers.ExtraHeaders {
		switch h {
		case "X-Auth-Timestamp":
			c.Set(h, fmt.Sprintf("%d", time.Now().Unix()))
		case "X-Auth-Route":
			host := c.Get("X-Forwarded-Host")
			uri := c.Get("X-Forwarded-Uri")
			c.Set(h, sanitizeHeaderValue(host+uri))
		}
	}
}

func setJWTMetadataHeaders(c *fiber.Ctx, cfg *config.Config, result *auth.AuthResult) {
	if !cfg.Headers.IncludeJWTMetadata || result.Metadata == nil {
		return
	}

	for k, v := range result.Metadata {
		// 首字母大写
		headerName := "X-Auth-" + strings.ToUpper(k[:1]) + k[1:]
		c.Set(headerName, sanitizeHeaderValue(v))
	}
}

func setInjectedAuthorization(c *fiber.Ctx, policy *config.RoutePolicy) {
	if policy == nil || policy.InjectAuthorization == "" {
		return
	}

	// 清理并限制长度，防止超长 header 导致 HTTP 431
	sanitized := sanitizeHeaderValue(policy.InjectAuthorization)
	c.Set("Authorization", sanitized)
}

// UnauthorizedResponse 返回认证失败响应
func UnauthorizedResponse(c *fiber.Ctx, cfg *config.Config, message string) error {
	// 设置 WWW-Authenticate headers
	authenticateMethods := []string{}

	if len(cfg.BasicAuths) > 0 {
		authenticateMethods = append(authenticateMethods, `Basic realm="api"`)
	}

	if len(cfg.BearerTokens) > 0 || cfg.JWT.Secret != "" {
		authenticateMethods = append(authenticateMethods, `Bearer realm="api"`)
	}

	// 添加所有 WWW-Authenticate headers
	for _, method := range authenticateMethods {
		c.Append("WWW-Authenticate", method)
	}

	// 设置缓存控制
	c.Set("Cache-Control", "no-store")

	// 返回 JSON 错误响应
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":     message,
		"timestamp": time.Now().Unix(),
	})
}

// sanitizeHeaderValue 清理 header 值，防止 header 注入攻击
func sanitizeHeaderValue(value string) string {
	// 移除换行符
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")

	// 限制长度
	if len(value) > 1024 {
		value = value[:1024]
	}

	return value
}
