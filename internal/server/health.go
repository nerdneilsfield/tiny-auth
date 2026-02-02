package server

import (
	"github.com/gofiber/fiber/v2"
)

// HandleHealth 处理健康检查请求
func (s *Server) HandleHealth(c *fiber.Ctx) error {
	cfg := s.GetConfig()

	return c.JSON(fiber.Map{
		"status":       "ok",
		"basic_count":  len(cfg.BasicAuths),
		"bearer_count": len(cfg.BearerTokens),
		"apikey_count": len(cfg.APIKeys),
		"jwt_enabled":  cfg.JWT.Secret != "",
		"policy_count": len(cfg.RoutePolicies),
	})
}

// HandleDebug 处理调试端点（显示配置摘要）
func (s *Server) HandleDebug(c *fiber.Ctx) error {
	cfg := s.GetConfig()

	// 构建安全的配置摘要（不包含敏感信息）
	basicNames := make([]string, 0, len(cfg.BasicAuths))
	for _, b := range cfg.BasicAuths {
		basicNames = append(basicNames, b.Name)
	}

	bearerNames := make([]string, 0, len(cfg.BearerTokens))
	for _, b := range cfg.BearerTokens {
		bearerNames = append(bearerNames, b.Name)
	}

	apiKeyNames := make([]string, 0, len(cfg.APIKeys))
	for _, k := range cfg.APIKeys {
		apiKeyNames = append(apiKeyNames, k.Name)
	}

	policyNames := make([]string, 0, len(cfg.RoutePolicies))
	for i := range cfg.RoutePolicies {
		policyNames = append(policyNames, cfg.RoutePolicies[i].Name)
	}

	return c.JSON(fiber.Map{
		"server": fiber.Map{
			"port":        cfg.Server.Port,
			"auth_path":   cfg.Server.AuthPath,
			"health_path": cfg.Server.HealthPath,
		},
		"authentication": fiber.Map{
			"basic_auth":    basicNames,
			"bearer_tokens": bearerNames,
			"api_keys":      apiKeyNames,
			"jwt_enabled":   cfg.JWT.Secret != "",
		},
		"policies": policyNames,
	})
}
