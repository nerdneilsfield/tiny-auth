package config

// Config 是 tiny-auth 的主配置结构
type Config struct {
	Server        ServerConfig      `toml:"server"`
	Headers       HeadersConfig     `toml:"headers"`
	Logging       LoggingConfig     `toml:"logging"`
	BasicAuths    []BasicAuthConfig `toml:"basic_auth"`
	BearerTokens  []BearerConfig    `toml:"bearer_token"`
	APIKeys       []APIKeyConfig    `toml:"api_key"`
	JWT           JWTConfig         `toml:"jwt"`
	RoutePolicies []RoutePolicy     `toml:"route_policy"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port           string   `toml:"port"`            // 监听端口
	AuthPath       string   `toml:"auth_path"`       // ForwardAuth 端点路径
	HealthPath     string   `toml:"health_path"`     // 健康检查端点路径
	ReadTimeout    int      `toml:"read_timeout"`    // 读超时（秒）
	WriteTimeout   int      `toml:"write_timeout"`   // 写超时（秒）
	TrustedProxies []string `toml:"trusted_proxies"` // 可信代理 IP/CIDR 列表（用于验证 X-Forwarded-* headers）
}

// HeadersConfig Header 配置
type HeadersConfig struct {
	UserHeader         string   `toml:"user_header"`          // 用户名 header
	RoleHeader         string   `toml:"role_header"`          // 角色 header
	MethodHeader       string   `toml:"method_header"`        // 认证方法 header
	ExtraHeaders       []string `toml:"extra_headers"`        // 额外的 headers
	IncludeJWTMetadata bool     `toml:"include_jwt_metadata"` // 是否包含 JWT 元数据
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Format string `toml:"format"` // 日志格式: "json" 或 "text"
	Level  string `toml:"level"`  // 日志级别: "debug", "info", "warn", "error"
}

// BasicAuthConfig Basic 认证配置
type BasicAuthConfig struct {
	Name  string   `toml:"name"`  // 唯一标识符
	User  string   `toml:"user"`  // 用户名
	Pass  string   `toml:"pass"`  // 密码（支持 env:VAR 语法）
	Roles []string `toml:"roles"` // 关联的角色
}

// BearerConfig Bearer Token 配置
type BearerConfig struct {
	Name  string   `toml:"name"`  // 唯一标识符
	Token string   `toml:"token"` // Token 值（支持 env:VAR 语法）
	Roles []string `toml:"roles"` // 关联的角色
}

// APIKeyConfig API Key 配置
type APIKeyConfig struct {
	Name  string   `toml:"name"`  // 唯一标识符
	Key   string   `toml:"key"`   // API Key 值（支持 env:VAR 语法）
	Roles []string `toml:"roles"` // 关联的角色
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `toml:"secret"`         // HS256 签名密钥（支持 env:VAR 语法）
	Issuer        string `toml:"issuer"`         // 期望的 issuer (iss claim)
	Audience      string `toml:"audience"`       // 期望的 audience (aud claim)
	UserClaimName string `toml:"user_claim_name"` // 用户标识的 claim 名称（默认为 "sub"，可配置为 "preferred_username" 等）
}

// RoutePolicy 路由策略配置
type RoutePolicy struct {
	Name                string   `toml:"name"`                  // 唯一标识符
	Host                string   `toml:"host"`                  // Host 匹配模式
	PathPrefix          string   `toml:"path_prefix"`           // 路径前缀
	Method              string   `toml:"method"`                // HTTP 方法
	AllowAnonymous      bool     `toml:"allow_anonymous"`       // 是否允许匿名访问
	AllowedBasicNames   []string `toml:"allowed_basic_names"`   // 允许的 Basic Auth 名称
	AllowedBearerNames  []string `toml:"allowed_bearer_names"`  // 允许的 Bearer Token 名称
	AllowedAPIKeyNames  []string `toml:"allowed_api_key_names"` // 允许的 API Key 名称
	JWTOnly             bool     `toml:"jwt_only"`              // 仅允许 JWT
	RequireAllRoles     []string `toml:"require_all_roles"`     // 必须拥有所有角色
	RequireAnyRole      []string `toml:"require_any_role"`      // 必须拥有任意一个角色
	InjectAuthorization string   `toml:"inject_authorization"`  // 注入的 Authorization header
}
