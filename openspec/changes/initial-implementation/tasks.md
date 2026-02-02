# Implementation Tasks

## Phase 1: Project Setup & Configuration

### Task 1.1: Update Project Dependencies
- [ ] Update `go.mod` with correct module name (`github.com/nerdneilsfield/tiny-auth`)
- [ ] Add dependencies:
  - `github.com/gofiber/fiber/v2` (HTTP framework)
  - `github.com/BurntSushi/toml` (config parsing)
  - `github.com/golang-jwt/jwt/v5` (JWT validation)
  - `github.com/spf13/cobra` (CLI framework)
  - `go.uber.org/zap` (logging)
- [ ] Run `go mod tidy`
- [ ] Update `main.go` to use new module path

### Task 1.2: Create Configuration Structures
- [ ] Create `internal/config/types.go`
  - Define `Config` struct
  - Define `ServerConfig` struct
  - Define `HeadersConfig` struct
  - Define `BasicAuthConfig`, `BearerConfig`, `APIKeyConfig` structs
  - Define `JWTConfig` struct
  - Define `RoutePolicy` struct
- [ ] Add TOML struct tags to all fields
- [ ] Document each field with comments (Chinese)

### Task 1.3: Implement Configuration Loader
- [ ] Create `internal/config/loader.go`
  - Function `LoadConfig(path string) (*Config, error)`
  - Use `toml.DecodeFile()` to parse file
  - Handle file not found errors
  - Support `CONFIG_PATH` environment variable
- [ ] Create `internal/config/defaults.go`
  - Function `ApplyDefaults(cfg *Config)`
  - Set default values for optional fields
  - Default roles: Basic=["user"], Bearer=["service"], API Key=["api"]

### Task 1.4: Implement Configuration Validator
- [ ] Create `internal/config/validator.go`
  - Function `Validate(cfg *Config) error`
  - Check required fields are present
  - Check for duplicate names (basic/bearer/apikey)
  - Check for duplicate credentials (users/tokens/keys)
  - Validate header names (regex: `^[A-Za-z][A-Za-z0-9-]*$`)
  - Validate JWT secret length (>= 32 chars)
  - Validate policy references (names exist)
  - Check for conflicting policy settings
- [ ] Return detailed validation errors with field path

### Task 1.5: Create Example Configuration
- [ ] Create `config.example.toml` in project root
  - Include all configuration sections with examples
  - Add comments explaining each field
  - Include common use cases (admin, API, webhook, public)

---

## Phase 2: Authentication Implementation

### Task 2.1: Define Authentication Types
- [ ] Create `internal/auth/types.go`
  - Define `AuthResult` struct
  - Define `AuthStore` struct with maps
  - Add helper method `NewAuthStore()` constructor

### Task 2.2: Build Authentication Store
- [ ] Create `internal/auth/store.go`
  - Function `BuildStore(cfg *Config) *AuthStore`
  - Build `BasicByUser` and `BasicByName` maps
  - Build `BearerByToken` and `BearerByName` maps
  - Build `APIKeyByKey` and `APIKeyByName` maps
  - Use indexed maps for O(1) lookups

### Task 2.3: Implement Basic Auth Handler
- [ ] Create `internal/auth/basic.go`
  - Function `TryBasic(authHeader string, store *AuthStore) *AuthResult`
  - Check for "Basic " prefix
  - Base64 decode credentials
  - Parse `username:password`
  - Look up user in `BasicByUser` map
  - Compare password using `crypto/subtle.ConstantTimeCompare`
  - Return `AuthResult` with user and roles

### Task 2.4: Implement Bearer Token Handler
- [ ] Create `internal/auth/bearer.go`
  - Function `TryBearer(authHeader string, store *AuthStore) *AuthResult`
  - Check for "Bearer " prefix
  - Extract token
  - Look up token in `BearerByToken` map
  - Use constant-time comparison
  - Return `AuthResult` with name and roles

### Task 2.5: Implement API Key Handler
- [ ] Create `internal/auth/apikey.go`
  - Function `TryAPIKeyAuth(authHeader string, store *AuthStore) *AuthResult`
    - Check for "ApiKey " prefix
    - Look up key in `APIKeyByKey` map
  - Function `TryAPIKeyHeader(headerValue string, store *AuthStore) *AuthResult`
    - Look up key directly (from `X-Api-Key` header)
  - Use constant-time comparison for both
  - Return `AuthResult` with name and roles

### Task 2.6: Implement JWT Handler
- [ ] Create `internal/auth/jwt.go`
  - Function `TryJWT(tokenString string, jwtCfg *JWTConfig) *AuthResult`
  - Parse JWT using `jwt.Parse()`
  - Validate HS256 signing method
  - Verify signature with secret
  - Check issuer (`iss` claim) if configured
  - Check audience (`aud` claim) if configured
  - Extract subject (`sub`) as username
  - Extract role claim (if present)
  - Return `AuthResult` with user, roles, and metadata (issuer)

### Task 2.7: Unit Tests for Authentication
- [ ] Test Basic Auth with valid credentials
- [ ] Test Basic Auth with invalid password
- [ ] Test Basic Auth with unknown user
- [ ] Test Bearer Token with valid token
- [ ] Test Bearer Token with invalid token
- [ ] Test API Key (Authorization header)
- [ ] Test API Key (X-Api-Key header)
- [ ] Test JWT with valid signature
- [ ] Test JWT with invalid signature
- [ ] Test JWT with wrong issuer
- [ ] Test JWT vs static Bearer token (priority)

---

## Phase 3: Policy Implementation

### Task 3.1: Implement Route Policy Matcher
- [ ] Create `internal/policy/matcher.go`
  - Function `MatchPolicy(policies []RoutePolicy, host, uri, method string) *RoutePolicy`
  - Implement host matching (exact and wildcard `*.`)
  - Implement path prefix matching
  - Implement HTTP method matching (case-insensitive)
  - Return first matching policy or `nil`

### Task 3.2: Implement Policy Checker
- [ ] Create `internal/policy/checker.go`
  - Function `CheckPolicy(policy *RoutePolicy, result *AuthResult, store *AuthStore) bool`
  - Check authentication method whitelist (allowed_*_names)
  - Check `jwt_only` flag
  - Check role requirements (`require_all_roles`)
  - Check role requirements (`require_any_role`)
  - Return `true` if policy satisfied, `false` otherwise

### Task 3.3: Unit Tests for Policy
- [ ] Test exact host matching
- [ ] Test wildcard host matching (`*.example.com`)
- [ ] Test path prefix matching
- [ ] Test HTTP method matching
- [ ] Test combined matching (host + path + method)
- [ ] Test anonymous access policy
- [ ] Test authentication method restrictions
- [ ] Test role requirements (all)
- [ ] Test role requirements (any)
- [ ] Test JWT only policy
- [ ] Test first-match-wins behavior

---

## Phase 4: HTTP Server Implementation

### Task 4.1: Create Server Setup
- [ ] Create `internal/server/server.go`
  - Function `NewServer(cfg *Config, store *AuthStore) *fiber.App`
  - Initialize Fiber with timeouts
  - Add logger middleware
  - Register routes: `/auth`, `/health`, `/debug/config`
  - Return Fiber app instance

### Task 4.2: Implement ForwardAuth Handler
- [ ] Create `internal/server/handler.go`
  - Function `HandleAuth(c *fiber.Ctx, cfg *Config, store *AuthStore) error`
  - Extract Traefik forwarded headers:
    - `X-Forwarded-Host`
    - `X-Forwarded-Uri`
    - `X-Forwarded-Method`
    - `X-Forwarded-For`
  - Match route policy
  - Check anonymous access
  - Try authentication methods (JWT → Bearer → Basic → API Key)
  - Check policy constraints
  - Return success or unauthorized response

### Task 4.3: Implement Response Helpers
- [ ] Create `internal/server/response.go`
  - Function `SuccessResponse(c *fiber.Ctx, cfg *Config, result *AuthResult) error`
    - Set status 200
    - Inject headers (user, role, method)
    - Inject custom headers (timestamp, route)
    - Inject authorization override (if policy specifies)
    - Inject JWT metadata (if enabled)
    - Return "ok" body
  - Function `UnauthorizedResponse(c *fiber.Ctx, cfg *Config) error`
    - Set status 401
    - Add `WWW-Authenticate` headers
    - Return JSON error response

### Task 4.4: Implement Health Endpoint
- [ ] Create `internal/server/health.go`
  - Function `HandleHealth(c *fiber.Ctx, cfg *Config) error`
  - Return JSON with status and config summary
  - Include counts: basic_auth, bearer_token, api_key, route_policy

### Task 4.5: Header Injection Logic
- [ ] Implement header sanitization (remove newlines, limit length)
- [ ] Set standard headers (X-Auth-User, X-Auth-Role, X-Auth-Method)
- [ ] Set custom headers (X-Auth-Timestamp, X-Auth-Route)
- [ ] Set JWT metadata headers (if enabled)
- [ ] Override Authorization header (if policy specifies)

### Task 4.6: Integration Tests
- [ ] Test `/auth` endpoint with valid Basic Auth
- [ ] Test `/auth` endpoint with invalid credentials
- [ ] Test `/auth` endpoint with Bearer token
- [ ] Test `/auth` endpoint with API Key
- [ ] Test `/auth` endpoint with JWT
- [ ] Test anonymous access route
- [ ] Test policy enforcement (role requirements)
- [ ] Test policy enforcement (method restrictions)
- [ ] Test header injection
- [ ] Test authorization transformation
- [ ] Test `/health` endpoint

---

## Phase 5: CLI Commands

### Task 5.1: Update Root Command
- [ ] Update `cmd/root.go`
  - Change app name to "tiny-auth"
  - Update description
  - Keep verbose flag
  - Keep logger integration

### Task 5.2: Create Server Command
- [ ] Create `cmd/server.go`
  - Command name: "server"
  - Flag: `--config` (default: "config.toml")
  - Load configuration
  - Validate configuration
  - Build auth store
  - Create Fiber server
  - Start server
  - Handle graceful shutdown (SIGTERM, SIGINT)
  - Log server start info (port, auth count, policy count)

### Task 5.3: Create Validate Command
- [ ] Create `cmd/validate.go`
  - Command name: "validate"
  - Argument: config file path (optional, default: "config.toml")
  - Load configuration
  - Run validation
  - Check file permissions (warn if > 0600)
  - Check for weak passwords (< 12 chars)
  - Print validation results with checkmarks/warnings
  - Exit with code 0 (success) or 1 (failure)

### Task 5.4: Update Version Command
- [ ] Update `cmd/version.go`
  - Display: version, build time, git commit
  - Keep existing implementation

### Task 5.5: Update Main Entry Point
- [ ] Update `main.go`
  - Change import paths to `github.com/nerdneilsfield/tiny-auth/cmd`
  - Keep signal handling for graceful shutdown
  - Keep logger initialization

---

## Phase 6: Documentation & Examples

### Task 6.1: Update README.md
- [ ] Project overview and purpose
- [ ] Features list
- [ ] Quick start guide
- [ ] Installation instructions
- [ ] Configuration reference
- [ ] Traefik integration examples (Docker labels)
- [ ] Usage examples (curl commands)
- [ ] Security best practices
- [ ] Troubleshooting section

### Task 6.2: Create Docker Support
- [ ] Create `Dockerfile`
  - Multi-stage build (builder + alpine)
  - Include ca-certificates
  - Set working directory
  - Copy binary and example config
  - Expose port 8080
  - Set CMD to run server
- [ ] Create `.dockerignore`

### Task 6.3: Create Docker Compose Example
- [ ] Create `docker-compose.yml`
  - Service: tiny-auth
  - Service: traefik (with ForwardAuth middleware)
  - Service: whoami (test upstream)
  - Volume: config.toml
  - Networks
  - Labels for Traefik integration

### Task 6.4: Create Configuration Documentation
- [ ] Create `docs/configuration.md`
  - Document all configuration sections
  - Document all fields with types and defaults
  - Provide examples for common scenarios
  - Document validation rules

### Task 6.5: Create Traefik Integration Guide
- [ ] Create `docs/traefik-integration.md`
  - Explain ForwardAuth mechanism
  - Provide Docker labels examples
  - Provide File provider examples
  - Explain `authResponseHeaders` setting
  - Explain `trustForwardHeader` setting
  - Document SSE/WebSocket considerations (forwardBody)

---

## Phase 7: Testing & Quality

### Task 7.1: Unit Test Coverage
- [ ] Achieve >80% coverage for `internal/auth/`
- [ ] Achieve >80% coverage for `internal/policy/`
- [ ] Achieve >80% coverage for `internal/config/`

### Task 7.2: Integration Test Suite
- [ ] Create `test/integration_test.go`
- [ ] Test full authentication flow with Fiber server
- [ ] Test multiple concurrent requests
- [ ] Test graceful shutdown

### Task 7.3: Benchmarks
- [ ] Create `test/benchmark_test.go`
- [ ] Benchmark Basic Auth lookup
- [ ] Benchmark Bearer Token lookup
- [ ] Benchmark JWT validation
- [ ] Benchmark policy matching
- [ ] Target: >1000 req/s on single core

### Task 7.4: Linting & Formatting
- [ ] Run `golangci-lint run`
- [ ] Fix all linting errors
- [ ] Run `gofmt -s -w .`
- [ ] Ensure all code is formatted

---

## Phase 8: Release Preparation

### Task 8.1: Update GoReleaser Config
- [ ] Update `.goreleaser.yml`
  - Correct binary name: `tiny-auth`
  - Update module path
  - Configure builds for linux/amd64, darwin/amd64, darwin/arm64
  - Configure archives
  - Configure Docker images

### Task 8.2: Create Example Configurations
- [ ] `examples/basic-auth.toml` - Simple Basic Auth
- [ ] `examples/jwt-only.toml` - JWT validation only
- [ ] `examples/multi-auth.toml` - All auth methods
- [ ] `examples/route-policies.toml` - Complex route policies
- [ ] `examples/production.toml` - Production-ready config

### Task 8.3: Final Testing
- [ ] Test with real Traefik setup (Docker Compose)
- [ ] Test all authentication methods end-to-end
- [ ] Test all route policy scenarios
- [ ] Test graceful shutdown
- [ ] Test configuration validation
- [ ] Test with invalid configurations

### Task 8.4: Version Tag
- [ ] Update version in `cmd/version.go` (will be overridden by build)
- [ ] Create git tag: `v0.1.0`
- [ ] Push tag to trigger GoReleaser

---

## Checklist Summary

**Phase 1: Project Setup** (5 tasks)
- Setup dependencies and configuration structures

**Phase 2: Authentication** (7 tasks)
- Implement all authentication methods with tests

**Phase 3: Policy** (3 tasks)
- Implement route matching and policy enforcement

**Phase 4: HTTP Server** (6 tasks)
- Build Fiber server with ForwardAuth handler

**Phase 5: CLI** (5 tasks)
- Create CLI commands (server, validate)

**Phase 6: Documentation** (5 tasks)
- Write comprehensive documentation and examples

**Phase 7: Testing** (4 tasks)
- Unit tests, integration tests, benchmarks

**Phase 8: Release** (4 tasks)
- Final testing and release preparation

**Total: 39 tasks**

---

## Progress Tracking

Use this section to track completed tasks:

- [ ] Phase 1 complete (0/5)
- [ ] Phase 2 complete (0/7)
- [ ] Phase 3 complete (0/3)
- [ ] Phase 4 complete (0/6)
- [ ] Phase 5 complete (0/5)
- [ ] Phase 6 complete (0/5)
- [ ] Phase 7 complete (0/4)
- [ ] Phase 8 complete (0/4)

**Overall Progress: 0/39 tasks complete (0%)**
