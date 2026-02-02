package server

import (
	"testing"

	"github.com/nerdneilsfield/tiny-auth/internal/auth"
	"github.com/nerdneilsfield/tiny-auth/internal/config"
)

func TestServerReload_RateLimiter(t *testing.T) {
	baseCfg := &config.Config{
		Server: config.ServerConfig{
			Port:         "3000",
			AuthPath:     "/auth",
			HealthPath:   "/health",
			ReadTimeout:  5,
			WriteTimeout: 5,
		},
		RateLimit: config.RateLimitConfig{
			Enabled:     false,
			MaxAttempts: 2,
			WindowSecs:  1,
			BanSecs:     1,
		},
	}

	srv := createTestServer(baseCfg)
	if srv.RateLimiter != nil {
		t.Fatal("expected rate limiter to be nil when disabled")
	}

	enabledCfg := &config.Config{
		Server: baseCfg.Server,
		RateLimit: config.RateLimitConfig{
			Enabled:     true,
			MaxAttempts: 2,
			WindowSecs:  1,
			BanSecs:     1,
		},
	}
	srv.Reload(enabledCfg, auth.BuildStore(enabledCfg))
	if srv.RateLimiter == nil {
		t.Fatal("expected rate limiter to be initialized after reload")
	}

	disabledCfg := &config.Config{
		Server: baseCfg.Server,
		RateLimit: config.RateLimitConfig{
			Enabled:     false,
			MaxAttempts: 2,
			WindowSecs:  1,
			BanSecs:     1,
		},
	}
	srv.Reload(disabledCfg, auth.BuildStore(disabledCfg))
	if srv.RateLimiter != nil {
		t.Fatal("expected rate limiter to be nil after disabling")
	}
}
