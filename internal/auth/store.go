package auth

import "github.com/nerdneilsfield/tiny-auth/internal/config"

// BuildStore 从配置构建认证存储
func BuildStore(cfg *config.Config) *AuthStore {
	store := NewAuthStore()

	// 构建 Basic Auth 索引
	for _, b := range cfg.BasicAuths {
		store.BasicByUser[b.User] = b
		store.BasicByName[b.Name] = b
	}

	// 构建 Bearer Token 索引
	for _, b := range cfg.BearerTokens {
		store.BearerByToken[b.Token] = b
		store.BearerByName[b.Name] = b
	}

	// 构建 API Key 索引
	for _, k := range cfg.APIKeys {
		store.APIKeyByKey[k.Key] = k
		store.APIKeyByName[k.Name] = k
	}

	return store
}
