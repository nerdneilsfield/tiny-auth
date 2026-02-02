package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

// Server 封装 Fiber 应用和配置
type Server struct {
	App    *fiber.App
	Config *config.Config
	Store  *auth.AuthStore
	mu     sync.RWMutex // 用于配置热重载时的并发控制
}

// NewServer 创建新的 HTTP 服务器
func NewServer(cfg *config.Config, store *auth.AuthStore) *Server {
	srv := &Server{
		Config: cfg,
		Store:  store,
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		ReadTimeout:           time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:          time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:           120 * time.Second,
		ServerHeader:          "tiny-auth",
		AppName:               "tiny-auth",
	})

	// 添加 recover 中间件（防止 panic 导致服务器崩溃）
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 添加日志中间件
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} ${method} ${path} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
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

	fmt.Printf("🔐 tiny-auth starting on :%s\n", port)
	fmt.Printf("   Auth endpoint: %s\n", s.Config.Server.AuthPath)
	fmt.Printf("   Health endpoint: %s\n", s.Config.Server.HealthPath)
	fmt.Printf("   Basic Auth: %d users\n", len(s.Config.BasicAuths))
	fmt.Printf("   Bearer Tokens: %d\n", len(s.Config.BearerTokens))
	fmt.Printf("   API Keys: %d\n", len(s.Config.APIKeys))
	if s.Config.JWT.Secret != "" {
		fmt.Printf("   JWT: enabled\n")
	}
	fmt.Printf("   Route Policies: %d\n", len(s.Config.RoutePolicies))

	return s.App.Listen(":" + port)
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() error {
	fmt.Println("🛑 Shutting down server...")
	return s.App.Shutdown()
}

// Reload 重新加载配置（用于热重载）
func (s *Server) Reload(cfg *config.Config, store *auth.AuthStore) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Config = cfg
	s.Store = store

	fmt.Println("♻️  Configuration reloaded")
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
