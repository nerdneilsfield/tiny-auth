package server

import (
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
	"go.uber.org/zap"
)

// Server 封装 Fiber 应用和配置
type Server struct {
	App          *fiber.App
	Config       *config.Config
	Store        *auth.AuthStore
	Logger       *zap.Logger
	trustedCIDRs []*net.IPNet // 可信代理 CIDR 列表（解析后）
	mu           sync.RWMutex // 用于配置热重载时的并发控制
}

// NewServer 创建新的 HTTP 服务器
func NewServer(cfg *config.Config, store *auth.AuthStore, logger *zap.Logger) *Server {
	// 解析可信代理配置
	trustedCIDRs := parseTrustedProxies(cfg.Server.TrustedProxies)
	if len(trustedCIDRs) > 0 {
		logger.Info("trusted proxies configured",
			zap.Int("count", len(trustedCIDRs)),
			zap.Strings("proxies", cfg.Server.TrustedProxies),
		)
	} else {
		logger.Warn("no trusted proxies configured - X-Forwarded-* headers accepted from ANY source",
			zap.String("recommendation", "set server.trusted_proxies for production"),
		)
	}

	srv := &Server{
		Config:       cfg,
		Store:        store,
		Logger:       logger,
		trustedCIDRs: trustedCIDRs,
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true, // 我们用自己的日志
		ReadTimeout:           time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:          time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:           120 * time.Second,
		ServerHeader:          "tiny-auth",
		AppName:               "tiny-auth",
	})

	// 添加 RequestID 中间件（用于追踪请求）
	app.Use(requestid.New())

	// 添加 recover 中间件（防止 panic 导致服务器崩溃）
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 注册路由
	app.All(cfg.Server.AuthPath, func(c *fiber.Ctx) error {
		return srv.HandleAuth(c)
	})

	app.Get(cfg.Server.HealthPath, func(c *fiber.Ctx) error {
		return srv.HandleHealth(c)
	})

	// 调试端点（可选）
	app.Get("/debug/config", func(c *fiber.Ctx) error {
		return srv.HandleDebug(c)
	})

	srv.App = app
	return srv
}

// Start 启动服务器
func (s *Server) Start() error {
	port := s.Config.Server.Port

	s.Logger.Info("tiny-auth starting",
		zap.String("port", port),
		zap.String("auth_endpoint", s.Config.Server.AuthPath),
		zap.String("health_endpoint", s.Config.Server.HealthPath),
		zap.Int("basic_auth_users", len(s.Config.BasicAuths)),
		zap.Int("bearer_tokens", len(s.Config.BearerTokens)),
		zap.Int("api_keys", len(s.Config.APIKeys)),
		zap.Bool("jwt_enabled", s.Config.JWT.Secret != ""),
		zap.Int("route_policies", len(s.Config.RoutePolicies)),
	)

	return s.App.Listen(":" + port)
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() error {
	s.Logger.Info("shutting down server")
	return s.App.Shutdown()
}

// Reload 重新加载配置（用于热重载）
func (s *Server) Reload(cfg *config.Config, store *auth.AuthStore) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 重新解析可信代理
	s.trustedCIDRs = parseTrustedProxies(cfg.Server.TrustedProxies)

	s.Config = cfg
	s.Store = store

	s.Logger.Info("configuration reloaded",
		zap.Int("basic_auth_users", len(cfg.BasicAuths)),
		zap.Int("bearer_tokens", len(cfg.BearerTokens)),
		zap.Int("api_keys", len(cfg.APIKeys)),
		zap.Bool("jwt_enabled", cfg.JWT.Secret != ""),
		zap.Int("route_policies", len(cfg.RoutePolicies)),
		zap.Int("trusted_proxies", len(s.trustedCIDRs)),
	)
}

// GetConfig 获取当前配置（线程安全）
func (s *Server) GetConfig() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Config
}

// GetStore 获取当前认证存储（线程安全）
func (s *Server) GetStore() *auth.AuthStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Store
}
