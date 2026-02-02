# Design: tiny-auth Implementation

## Architecture Overview

### System Context

```
┌──────────┐         ┌─────────────┐         ┌──────────────┐
│  Client  │────────▶│   Traefik   │────────▶│  tiny-auth   │
└──────────┘         │   (Proxy)   │         │   (Auth)     │
                     └─────────────┘         └──────────────┘
                            │                        │
                            │                        │
                            │  ◄─ 200 + Headers      │
                            │  ◄─ 401 Unauthorized   │
                            ▼                        │
                     ┌─────────────┐                 │
                     │  Upstream   │◄────────────────┘
                     │   Service   │
                     └─────────────┘
```

### Component Architecture

```
tiny-auth/
├── main.go                    # Entry point, signal handling
├── cmd/
│   ├── root.go                # Cobra root command
│   ├── server.go              # Server start command
│   ├── validate.go            # Config validation command
│   └── version.go             # Version info command
├── internal/
│   ├── config/
│   │   ├── types.go           # Configuration structures
│   │   ├── loader.go          # TOML loading
│   │   ├── validator.go       # Validation logic
│   │   └── defaults.go        # Default values
│   ├── auth/
│   │   ├── types.go           # AuthResult, AuthStore types
│   │   ├── store.go           # Build indexed auth store
│   │   ├── basic.go           # Basic Auth handler
│   │   ├── bearer.go          # Bearer Token handler
│   │   ├── apikey.go          # API Key handler
│   │   └── jwt.go             # JWT handler
│   ├── policy/
│   │   ├── matcher.go         # Route policy matching
│   │   └── checker.go         # Policy enforcement
│   └── server/
│       ├── server.go          # Fiber server setup
│       ├── handler.go         # /auth endpoint handler
│       ├── health.go          # /health endpoint
│       └── response.go        # Response helpers
└── pkg/
    └── logger/
        └── logger.go          # Structured logging utilities
```

## Data Structures

### Configuration Types

```go
// internal/config/types.go

type Config struct {
    Server        ServerConfig        `toml:"server"`
    Headers       HeadersConfig       `toml:"headers"`
    BasicAuths    []BasicAuthConfig   `toml:"basic_auth"`
    BearerTokens  []BearerConfig      `toml:"bearer_token"`
    APIKeys       []APIKeyConfig      `toml:"api_key"`
    JWT           JWTConfig           `toml:"jwt"`
    RoutePolicies []RoutePolicy       `toml:"route_policy"`
}

type ServerConfig struct {
    Port         string `toml:"port"`
    AuthPath     string `toml:"auth_path"`
    HealthPath   string `toml:"health_path"`
    ReadTimeout  int    `toml:"read_timeout"`
    WriteTimeout int    `toml:"write_timeout"`
}

type HeadersConfig struct {
    UserHeader         string   `toml:"user_header"`
    RoleHeader         string   `toml:"role_header"`
    MethodHeader       string   `toml:"method_header"`
    ExtraHeaders       []string `toml:"extra_headers"`
    IncludeJWTMetadata bool     `toml:"include_jwt_metadata"`
}

type BasicAuthConfig struct {
    Name  string   `toml:"name"`
    User  string   `toml:"user"`
    Pass  string   `toml:"pass"`
    Roles []string `toml:"roles"`
}

type BearerConfig struct {
    Name  string   `toml:"name"`
    Token string   `toml:"token"`
    Roles []string `toml:"roles"`
}

type APIKeyConfig struct {
    Name  string   `toml:"name"`
    Key   string   `toml:"key"`
    Roles []string `toml:"roles"`
}

type JWTConfig struct {
    Secret   string `toml:"secret"`
    Issuer   string `toml:"issuer"`
    Audience string `toml:"audience"`
}

type RoutePolicy struct {
    Name                string   `toml:"name"`
    Host                string   `toml:"host"`
    PathPrefix          string   `toml:"path_prefix"`
    Method              string   `toml:"method"`
    AllowAnonymous      bool     `toml:"allow_anonymous"`
    AllowedBasicNames   []string `toml:"allowed_basic_names"`
    AllowedBearerNames  []string `toml:"allowed_bearer_names"`
    AllowedAPIKeyNames  []string `toml:"allowed_api_key_names"`
    JWTOnly             bool     `toml:"jwt_only"`
    RequireAllRoles     []string `toml:"require_all_roles"`
    RequireAnyRole      []string `toml:"require_any_role"`
    InjectAuthorization string   `toml:"inject_authorization"`
}
```

### Authentication Types

```go
// internal/auth/types.go

type AuthResult struct {
    Method   string            // "basic", "bearer", "apikey", "jwt", "anonymous"
    Name     string            // Config name (e.g., "admin-user")
    User     string            // Username or subject
    Roles    []string          // Associated roles
    Metadata map[string]string // Extra info (e.g., JWT issuer)
}

type AuthStore struct {
    // Fast lookups
    BasicByUser   map[string]BasicAuthConfig  // user -> config
    BearerByToken map[string]BearerConfig     // token -> config
    APIKeyByKey   map[string]APIKeyConfig     // key -> config
    
    // Name lookups (for policy validation)
    BasicByName   map[string]BasicAuthConfig
    BearerByName  map[string]BearerConfig
    APIKeyByName  map[string]APIKeyConfig
}
```

## Key Algorithms

### Authentication Flow

```go
func handleAuth(c *fiber.Ctx, cfg *Config, store *AuthStore) error {
    // 1. Extract Traefik forwarded headers
    originalHost := c.Get("X-Forwarded-Host", c.Get("X-Forwarded-Server"))
    originalURI := c.Get("X-Forwarded-Uri")
    originalMethod := c.Get("X-Forwarded-Method")
    
    // 2. Match route policy
    policy := matchPolicy(cfg.RoutePolicies, originalHost, originalURI, originalMethod)
    
    // 3. Check anonymous access
    if policy != nil && policy.AllowAnonymous {
        return successResponse(c, cfg, &AuthResult{Method: "anonymous"})
    }
    
    // 4. Try authentication methods (priority order)
    authHeader := c.Get("Authorization")
    var result *AuthResult
    
    // JWT (highest priority if configured and looks like JWT)
    if cfg.JWT.Secret != "" && strings.HasPrefix(authHeader, "Bearer ") {
        token := strings.TrimPrefix(authHeader, "Bearer ")
        if len(strings.Split(token, ".")) == 3 {
            result = tryJWT(token, &cfg.JWT)
        }
    }
    
    // Bearer Token (static)
    if result == nil && strings.HasPrefix(authHeader, "Bearer ") {
        result = tryBearer(authHeader, store)
    }
    
    // Basic Auth
    if result == nil && strings.HasPrefix(authHeader, "Basic ") {
        result = tryBasic(authHeader, store)
    }
    
    // API Key (Authorization: ApiKey)
    if result == nil && strings.HasPrefix(authHeader, "ApiKey ") {
        result = tryAPIKeyAuth(authHeader, store)
    }
    
    // API Key (X-Api-Key header)
    if result == nil {
        result = tryAPIKeyHeader(c.Get("X-Api-Key"), store)
    }
    
    // 5. Check policy constraints
    if result != nil && checkPolicy(policy, result, store) {
        return successResponse(c, cfg, result)
    }
    
    // 6. Authentication failed
    return unauthorizedResponse(c, cfg)
}
```

### Policy Matching

```go
func matchPolicy(policies []RoutePolicy, host, uri, method string) *RoutePolicy {
    for _, p := range policies {
        // Match host (exact or wildcard)
        if p.Host != "" {
            if strings.HasPrefix(p.Host, "*.") {
                suffix := p.Host[1:] // .example.com
                if !strings.HasSuffix(host, suffix) {
                    continue
                }
            } else if !strings.EqualFold(p.Host, host) {
                continue
            }
        }
        
        // Match path prefix
        if p.PathPrefix != "" && !strings.HasPrefix(uri, p.PathPrefix) {
            continue
        }
        
        // Match method
        if p.Method != "" && !strings.EqualFold(p.Method, method) {
            continue
        }
        
        // All criteria matched
        return &p
    }
    return nil
}
```

### Policy Checking

```go
func checkPolicy(policy *RoutePolicy, result *AuthResult, store *AuthStore) bool {
    if policy == nil {
        return true // No policy = accept any valid auth
    }
    
    // Check authentication method whitelist
    switch result.Method {
    case "basic":
        if len(policy.AllowedBasicNames) > 0 && !contains(policy.AllowedBasicNames, result.Name) {
            return false
        }
    case "bearer":
        if len(policy.AllowedBearerNames) > 0 && !contains(policy.AllowedBearerNames, result.Name) {
            return false
        }
    case "apikey":
        if len(policy.AllowedAPIKeyNames) > 0 && !contains(policy.AllowedAPIKeyNames, result.Name) {
            return false
        }
    case "jwt":
        // JWT allowed if jwt_only or no specific restrictions
    }
    
    // Check role requirements (require ALL)
    if len(policy.RequireAllRoles) > 0 {
        for _, required := range policy.RequireAllRoles {
            if !contains(result.Roles, required) {
                return false
            }
        }
    }
    
    // Check role requirements (require ANY)
    if len(policy.RequireAnyRole) > 0 {
        hasAny := false
        for _, required := range policy.RequireAnyRole {
            if contains(result.Roles, required) {
                hasAny = true
                break
            }
        }
        if !hasAny {
            return false
        }
    }
    
    return true
}
```

## Security Considerations

### 1. Constant-Time Comparison

**All credential comparisons must use `crypto/subtle.ConstantTimeCompare`:**

```go
// ✅ Correct
if subtle.ConstantTimeCompare([]byte(inputPass), []byte(storedPass)) == 1 {
    // Authenticated
}

// ❌ Wrong (timing attack vulnerable)
if inputPass == storedPass {
    // Authenticated
}
```

### 2. No Secrets in Logs

```go
// ✅ Correct
log.Info("Authentication successful",
    zap.String("user", result.User),
    zap.String("method", result.Method),
)

// ❌ Wrong
log.Info("Authentication successful",
    zap.String("password", inputPass), // Never log!
)
```

### 3. Secure HTTP Headers

```go
// Sanitize header values
func sanitizeHeaderValue(v string) string {
    // Remove newlines (prevent header injection)
    v = strings.ReplaceAll(v, "\r", "")
    v = strings.ReplaceAll(v, "\n", "")
    
    // Limit length
    if len(v) > 1024 {
        v = v[:1024]
    }
    
    return v
}
```

### 4. JWT Validation

```go
// Parse and validate JWT
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // Verify signing method
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return []byte(cfg.JWT.Secret), nil
})

if err != nil || !token.Valid {
    return nil // Invalid token
}

// Verify issuer/audience
claims := token.Claims.(jwt.MapClaims)
if cfg.JWT.Issuer != "" && claims["iss"] != cfg.JWT.Issuer {
    return nil
}
if cfg.JWT.Audience != "" && claims["aud"] != cfg.JWT.Audience {
    return nil
}
```

## Performance Optimization

### 1. Indexed Data Structures

**Build hash maps at startup for O(1) lookups:**

```go
func buildAuthStore(cfg *Config) *AuthStore {
    store := &AuthStore{
        BasicByUser:   make(map[string]BasicAuthConfig),
        BearerByToken: make(map[string]BearerConfig),
        APIKeyByKey:   make(map[string]APIKeyConfig),
        BasicByName:   make(map[string]BasicAuthConfig),
        BearerByName:  make(map[string]BearerConfig),
        APIKeyByName:  make(map[string]APIKeyConfig),
    }
    
    for _, b := range cfg.BasicAuths {
        store.BasicByUser[b.User] = b
        store.BasicByName[b.Name] = b
    }
    
    // ... similar for other auth types
    
    return store
}
```

### 2. Avoid Regex Where Possible

**Use string operations instead of regex:**

```go
// ✅ Fast
if strings.HasPrefix(uri, "/api") { ... }

// ❌ Slower (for simple prefix match)
if matched, _ := regexp.MatchString("^/api", uri); matched { ... }
```

### 3. Connection Reuse

```go
// Configure Fiber for performance
app := fiber.New(fiber.Config{
    DisableStartupMessage: false,
    ReadTimeout:           5 * time.Second,
    WriteTimeout:          5 * time.Second,
    IdleTimeout:           120 * time.Second,
    ReadBufferSize:        4096,
    WriteBufferSize:       4096,
})
```

## Error Handling

### HTTP Status Codes

| Code | Usage |
|------|-------|
| `200` | Authentication successful |
| `401` | Authentication failed (invalid/missing credentials) |
| `403` | Policy violation (valid auth but not allowed by policy) |
| `500` | Internal server error (shouldn't happen) |

**Note:** Currently we use `401` for both invalid auth and policy violations. Consider using `403` for policy violations in future.

### Response Format

**Success (200):**
```http
HTTP/1.1 200 OK
X-Auth-Method: basic
X-Auth-User: admin
X-Auth-Role: admin,user
Content-Length: 2

ok
```

**Failure (401):**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Basic realm="api"
WWW-Authenticate: Bearer realm="api"
Content-Type: application/json

{
  "error": "Unauthorized",
  "timestamp": 1738560000
}
```

## Logging Strategy

### Log Levels

- **INFO**: Successful auth, server start/stop, config loaded
- **WARN**: Configuration warnings, policy mismatches
- **ERROR**: Authentication failures, config errors
- **DEBUG**: Detailed auth flow (only in verbose mode)

### Log Format

```json
{
  "level": "info",
  "timestamp": "2026-02-03T10:00:00Z",
  "message": "authentication successful",
  "host": "api.example.com",
  "path": "/api/users",
  "method": "GET",
  "auth_method": "basic",
  "user": "admin",
  "roles": ["admin", "user"],
  "ip": "192.168.1.100"
}
```

**Never log:**
- Passwords
- Tokens
- API keys
- Full `Authorization` header

## Testing Strategy

### Unit Tests

- Each authentication method handler
- Policy matching logic
- Policy checking logic
- Configuration validation
- Header injection

### Integration Tests

- End-to-end auth flow with Fiber server
- Multiple authentication methods
- Route policy enforcement
- Header injection validation

### Test Doubles

```go
// Mock AuthStore for testing
func newTestAuthStore() *AuthStore {
    return &AuthStore{
        BasicByUser: map[string]BasicAuthConfig{
            "admin": {Name: "admin", User: "admin", Pass: "secret", Roles: []string{"admin"}},
        },
        BearerByToken: map[string]BearerConfig{
            "token123": {Name: "token1", Token: "token123", Roles: []string{"api"}},
        },
    }
}
```

## Deployment Considerations

### Docker Image

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o tiny-auth

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/tiny-auth .
COPY config.toml .
EXPOSE 8080
CMD ["./tiny-auth", "server"]
```

### Health Check

```bash
# Kubernetes/Docker health check
curl -f http://localhost:8080/health || exit 1
```

### Resource Requirements

**Minimal:**
- CPU: 0.1 cores
- Memory: 32 MB
- Disk: 10 MB

**Recommended:**
- CPU: 0.5 cores
- Memory: 128 MB
- Disk: 50 MB

## Future Enhancements

1. **Configuration Hot-Reload**: Watch config file for changes
2. **Prometheus Metrics**: `/metrics` endpoint for monitoring
3. **Redis Rate Limiting**: Distributed rate limiting
4. **LDAP Integration**: Authenticate against LDAP/AD
5. **mTLS**: Client certificate authentication
6. **OpenTelemetry**: Distributed tracing support
7. **Environment Variable Substitution**: `${ENV_VAR}` in config

## References

- [Fiber Documentation](https://docs.gofiber.io/)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [Traefik ForwardAuth](https://doc.traefik.io/traefik/middlewares/http/forwardauth/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
