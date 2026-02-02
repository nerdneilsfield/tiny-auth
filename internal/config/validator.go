package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var headerNameRegex = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*$`)

// Validate 验证配置
func Validate(cfg *Config) error {
	// 验证服务器配置
	if err := validateServer(&cfg.Server); err != nil {
		return fmt.Errorf("server: %w", err)
	}

	// 验证 Header 配置
	if err := validateHeaders(&cfg.Headers); err != nil {
		return fmt.Errorf("headers: %w", err)
	}

	// 验证日志配置
	if err := validateLogging(&cfg.Logging); err != nil {
		return fmt.Errorf("logging: %w", err)
	}

	// 验证 Basic Auth
	if err := validateBasicAuths(cfg.BasicAuths); err != nil {
		return fmt.Errorf("basic_auth: %w", err)
	}

	// 验证 Bearer Token
	if err := validateBearerTokens(cfg.BearerTokens); err != nil {
		return fmt.Errorf("bearer_token: %w", err)
	}

	// 验证 API Key
	if err := validateAPIKeys(cfg.APIKeys); err != nil {
		return fmt.Errorf("api_key: %w", err)
	}

	// 验证 JWT
	if err := validateJWT(&cfg.JWT); err != nil {
		return fmt.Errorf("jwt: %w", err)
	}

	// 验证路由策略
	if err := validateRoutePolicies(cfg.RoutePolicies, cfg); err != nil {
		return fmt.Errorf("route_policy: %w", err)
	}

	// 高级验证：循环依赖检测
	if err := validatePolicyDependencies(cfg); err != nil {
		return fmt.Errorf("policy dependencies: %w", err)
	}

	// 高级验证：策略冲突检测
	if err := validatePolicyConflicts(cfg.RoutePolicies); err != nil {
		return fmt.Errorf("policy conflicts: %w", err)
	}

	// 高级验证：JWT Secret 强度检测
	if err := validateJWTSecretStrength(&cfg.JWT); err != nil {
		return fmt.Errorf("jwt security: %w", err)
	}

	return nil
}

func validateServer(cfg *ServerConfig) error {
	// 验证端口
	if cfg.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	// 验证路径
	if !strings.HasPrefix(cfg.AuthPath, "/") {
		return fmt.Errorf("auth_path must start with /")
	}
	if !strings.HasPrefix(cfg.HealthPath, "/") {
		return fmt.Errorf("health_path must start with /")
	}

	// 验证超时
	if cfg.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}
	if cfg.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}

	return nil
}

func validateHeaders(cfg *HeadersConfig) error {
	headers := []string{cfg.UserHeader, cfg.RoleHeader, cfg.MethodHeader}
	headers = append(headers, cfg.ExtraHeaders...)

	seen := make(map[string]bool)
	for _, h := range headers {
		if h == "" {
			continue
		}

		// 验证 header 名称格式
		if !headerNameRegex.MatchString(h) {
			return fmt.Errorf("invalid header name %q (must match: ^[A-Za-z][A-Za-z0-9-]*$)", h)
		}

		// 检查重复
		lower := strings.ToLower(h)
		if seen[lower] {
			return fmt.Errorf("duplicate header name %q", h)
		}
		seen[lower] = true

		// 检查保留 headers
		if isReservedHeader(lower) {
			return fmt.Errorf("cannot use reserved header %q", h)
		}
	}

	return nil
}

func validateLogging(cfg *LoggingConfig) error {
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[cfg.Format] {
		return fmt.Errorf("format must be 'json' or 'text', got %q", cfg.Format)
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Level] {
		return fmt.Errorf("level must be 'debug', 'info', 'warn', or 'error', got %q", cfg.Level)
	}

	return nil
}

func validateBasicAuths(configs []BasicAuthConfig) error {
	if len(configs) == 0 {
		return nil // Basic Auth 是可选的
	}

	names := make(map[string]bool)
	users := make(map[string]bool)

	for _, cfg := range configs {
		// 验证必填字段
		if cfg.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if cfg.User == "" {
			return fmt.Errorf("[%s] user cannot be empty", cfg.Name)
		}
		// pass 和 pass_hash 至少要有一个
		if cfg.Pass == "" && cfg.PassHash == "" {
			return fmt.Errorf("[%s] either pass or pass_hash must be provided", cfg.Name)
		}

		// 如果同时提供了 pass 和 pass_hash，发出警告（优先使用 pass_hash）
		if cfg.Pass != "" && cfg.PassHash != "" {
			fmt.Fprintf(os.Stderr, "⚠ Warning: Basic auth [%s] has both pass and pass_hash configured. pass_hash will be used.\n", cfg.Name)
		}

		// 检查重复名称
		if names[cfg.Name] {
			return fmt.Errorf("duplicate name %q", cfg.Name)
		}
		names[cfg.Name] = true

		// 检查重复用户名
		if users[cfg.User] {
			return fmt.Errorf("duplicate user %q", cfg.User)
		}
		users[cfg.User] = true

		// 弱密码警告（仅对明文密码，bcrypt 哈希已经足够安全）
		if cfg.PassHash == "" && cfg.Pass != "" {
			if len(cfg.Pass) < 12 && !strings.HasPrefix(cfg.Pass, "env:") {
				fmt.Fprintf(os.Stderr, "⚠ Warning: Basic auth [%s] has short password (< 12 chars). Consider using pass_hash with bcrypt.\n", cfg.Name)
			}
		}
	}

	return nil
}

// secretConfig 定义具有 name 和 secret 字段的配置接口
type secretConfig interface {
	getName() string
	getSecret() string
}

// 为 BearerConfig 实现 secretConfig 接口
func (c BearerConfig) getName() string   { return c.Name }
func (c BearerConfig) getSecret() string { return c.Token }

// 为 APIKeyConfig 实现 secretConfig 接口
func (c APIKeyConfig) getName() string   { return c.Name }
func (c APIKeyConfig) getSecret() string { return c.Key }

// validateSecretConfigs 通用验证函数，使用泛型避免代码重复
func validateSecretConfigs[T secretConfig](configs []T, secretFieldName string) error {
	if len(configs) == 0 {
		return nil
	}

	names := make(map[string]bool)
	secrets := make(map[string]bool)

	for _, cfg := range configs {
		name := cfg.getName()
		secret := cfg.getSecret()

		// 验证 name 字段
		if name == "" {
			return fmt.Errorf("name cannot be empty")
		}

		// 验证 secret 字段
		if secret == "" {
			return fmt.Errorf("[%s] %s cannot be empty", name, secretFieldName)
		}

		// 检查重复 name
		if names[name] {
			return fmt.Errorf("duplicate name %q", name)
		}
		names[name] = true

		// 检查重复 secret
		if secrets[secret] {
			return fmt.Errorf("duplicate %s for name %q", secretFieldName, name)
		}
		secrets[secret] = true
	}

	return nil
}

// validateBearerTokens 使用通用验证函数
func validateBearerTokens(configs []BearerConfig) error {
	return validateSecretConfigs(configs, "token")
}

// validateAPIKeys 使用通用验证函数
func validateAPIKeys(configs []APIKeyConfig) error {
	return validateSecretConfigs(configs, "key")
}

func validateJWT(cfg *JWTConfig) error {
	if cfg.Secret == "" {
		return nil // JWT 是可选的
	}

	// 验证密钥长度（至少 256 bits = 32 bytes）
	if len(cfg.Secret) < 32 && !strings.HasPrefix(cfg.Secret, "env:") {
		return fmt.Errorf("secret must be at least 32 characters (256 bits), got %d", len(cfg.Secret))
	}

	return nil
}

func validateRoutePolicies(policies []RoutePolicy, cfg *Config) error {
	if len(policies) == 0 {
		return nil
	}

	names := make(map[string]bool)

	// 构建名称索引
	basicNames := make(map[string]bool)
	for _, b := range cfg.BasicAuths {
		basicNames[b.Name] = true
	}

	bearerNames := make(map[string]bool)
	for _, b := range cfg.BearerTokens {
		bearerNames[b.Name] = true
	}

	apiKeyNames := make(map[string]bool)
	for _, k := range cfg.APIKeys {
		apiKeyNames[k.Name] = true
	}

	for _, policy := range policies {
		if policy.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}

		// 检查重复名称
		if names[policy.Name] {
			return fmt.Errorf("duplicate name %q", policy.Name)
		}
		names[policy.Name] = true

		// 验证引用的认证名称
		for _, name := range policy.AllowedBasicNames {
			if !basicNames[name] {
				return fmt.Errorf("[%s] references unknown basic_auth %q", policy.Name, name)
			}
		}

		for _, name := range policy.AllowedBearerNames {
			if !bearerNames[name] {
				return fmt.Errorf("[%s] references unknown bearer_token %q", policy.Name, name)
			}
		}

		for _, name := range policy.AllowedAPIKeyNames {
			if !apiKeyNames[name] {
				return fmt.Errorf("[%s] references unknown api_key %q", policy.Name, name)
			}
		}

		// 警告：匿名访问与角色要求冲突
		if policy.AllowAnonymous && (len(policy.RequireAllRoles) > 0 || len(policy.RequireAnyRole) > 0) {
			fmt.Fprintf(os.Stderr, "⚠ Warning: Policy [%s] allows anonymous but requires roles (roles will be ignored)\n", policy.Name)
		}

		// 警告：JWT only 与其他方法限制冲突
		if policy.JWTOnly && (len(policy.AllowedBasicNames) > 0 || len(policy.AllowedBearerNames) > 0 || len(policy.AllowedAPIKeyNames) > 0) {
			fmt.Fprintf(os.Stderr, "⚠ Warning: Policy [%s] is jwt_only but has other method restrictions (will be ignored)\n", policy.Name)
		}
	}

	return nil
}

func isReservedHeader(name string) bool {
	reserved := []string{"host", "content-length", "transfer-encoding"}
	for _, r := range reserved {
		if name == r {
			return true
		}
	}
	return false
}
