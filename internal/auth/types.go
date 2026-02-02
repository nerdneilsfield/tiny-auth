package auth

import "github.com/nerdneilsfield/tiny-auth/internal/config"

// AuthResult 认证结果
//
//nolint:revive // exported name is stable API surface
type AuthResult struct {
	Method   string            // 认证方法: "basic", "bearer", "apikey", "jwt", "anonymous"
	Name     string            // 配置名称（如 "admin-user"）
	User     string            // 用户名或 subject
	Roles    []string          // 关联的角色
	Metadata map[string]string // 额外的元数据（如 JWT issuer）
}

// AuthStore 认证存储，用于快速查找
//
//nolint:revive // exported name is stable API surface
type AuthStore struct {
	// 按凭证查找（用于认证）
	BasicByUser   map[string]config.BasicAuthConfig
	BearerByToken map[string]config.BearerConfig
	APIKeyByKey   map[string]config.APIKeyConfig

	// 按名称查找（用于策略验证）
	BasicByName  map[string]config.BasicAuthConfig
	BearerByName map[string]config.BearerConfig
	APIKeyByName map[string]config.APIKeyConfig
}

// NewAuthStore 创建新的认证存储
func NewAuthStore() *AuthStore {
	return &AuthStore{
		BasicByUser:   make(map[string]config.BasicAuthConfig),
		BearerByToken: make(map[string]config.BearerConfig),
		APIKeyByKey:   make(map[string]config.APIKeyConfig),
		BasicByName:   make(map[string]config.BasicAuthConfig),
		BearerByName:  make(map[string]config.BearerConfig),
		APIKeyByName:  make(map[string]config.APIKeyConfig),
	}
}
