# Proposal: Initial Implementation of tiny-auth

## Summary

Implement a production-ready Traefik ForwardAuth authentication service with support for multiple authentication methods (Basic Auth, Bearer Token, API Key, JWT) and flexible route-based policy control.

## Motivation

### Problem

When deploying microservices behind Traefik, teams need:

1. **Centralized Authentication**: Avoid implementing auth in every service
2. **Multiple Auth Methods**: Different services/clients need different auth types
3. **Flexible Policies**: Different routes require different authentication strategies
4. **Header Transformation**: Client authentication may differ from upstream requirements (e.g., client uses Basic Auth, upstream needs Bearer token)

Current solutions are either:
- Too heavyweight (full OAuth/OIDC servers)
- Too limited (Traefik's BasicAuth middleware only supports one method)
- Too inflexible (can't do route-specific policies or header injection)

### Solution

Build tiny-auth as a lightweight, configuration-driven ForwardAuth service that:

1. **Validates** requests using configured authentication methods
2. **Matches** requests against route policies (host/path/method)
3. **Injects** appropriate headers for upstream services
4. **Returns** 200 (success) or 401/403 (failure) to Traefik

## Goals

### Core Functionality

1. **Multiple Authentication Methods**
   - Basic Auth (username/password)
   - Bearer Token (static tokens)
   - API Key (via header or Authorization)
   - JWT (with signature validation)

2. **Route-Based Policies**
   - Match by host (with wildcard support)
   - Match by path prefix
   - Match by HTTP method
   - Allow anonymous access for specific routes
   - Restrict to specific auth methods/credentials
   - Enforce role requirements

3. **Header Injection**
   - Inject user/role/method headers
   - Transform Authorization header for upstream
   - Add custom headers (timestamp, route info, etc.)

4. **Production Quality**
   - Configuration validation
   - Health check endpoint
   - Structured logging
   - Graceful shutdown
   - Security best practices (constant-time comparison)

### Non-Goals

- User management UI or database
- OAuth/OIDC provider functionality
- Built-in rate limiting (use Traefik's)
- Session management or cookies

## Design Overview

### Architecture

```
Request Flow:
Client → Traefik → tiny-auth (/auth) → [Validate] → Return 200/401
                       ↓                                    ↓
                   (if 200)                          Copy headers
                       ↓                                    ↓
                   Upstream ← ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
```

### Key Components

1. **Config Loader** (`internal/config/`)
   - Parse TOML configuration
   - Build indexed data structures for fast lookup
   - Validate configuration on startup

2. **Auth Handlers** (`internal/auth/`)
   - Basic Auth: base64 decode + constant-time compare
   - Bearer: token lookup in map
   - API Key: check `Authorization: ApiKey` or `X-Api-Key` header
   - JWT: parse, validate signature, check issuer/audience

3. **Policy Matcher** (`internal/policy/`)
   - Match route policies by host/path/method
   - Check if auth result satisfies policy requirements
   - Handle anonymous routes

4. **HTTP Server** (`internal/server/`)
   - Fiber-based HTTP server
   - `/auth` endpoint (Traefik ForwardAuth)
   - `/health` endpoint
   - `/debug/config` endpoint (optional)

5. **CLI Commands** (`cmd/`)
   - `server`: Start the auth service
   - `validate`: Validate configuration file
   - `version`: Show version info

### Configuration Example

```toml
[server]
port = "8080"
auth_path = "/auth"
health_path = "/health"

[headers]
user_header = "X-Auth-User"
role_header = "X-Auth-Role"
method_header = "X-Auth-Method"

[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "supersecret"
roles = ["admin", "user"]

[[bearer_token]]
name = "api-token"
token = "sk-prod-xxx"
roles = ["api", "service"]

[[route_policy]]
name = "admin-only"
host = "admin.example.com"
allowed_basic_names = ["admin-user"]
require_all_roles = ["admin"]

[[route_policy]]
name = "public-api"
host = "api.example.com"
path_prefix = "/public"
allow_anonymous = true
```

## Implementation Plan

### Phase 1: Core Structure (Day 1)

1. Update `go.mod` with dependencies
2. Create directory structure
3. Implement config structures and TOML loader
4. Add configuration validation

### Phase 2: Authentication (Day 1-2)

1. Implement Basic Auth handler
2. Implement Bearer Token handler
3. Implement API Key handler
4. Implement JWT handler
5. Create auth store with indexed lookups

### Phase 3: Policy & Routing (Day 2)

1. Implement route policy matcher
2. Implement policy checker (role enforcement)
3. Add wildcard host matching
4. Handle anonymous routes

### Phase 4: HTTP Server (Day 2-3)

1. Set up Fiber server
2. Implement `/auth` ForwardAuth handler
3. Add `/health` endpoint
4. Add logging middleware
5. Implement header injection logic

### Phase 5: CLI & Polish (Day 3)

1. Update cobra commands (server, validate)
2. Add comprehensive error messages
3. Create example configuration files
4. Write integration tests

### Phase 6: Documentation (Day 3-4)

1. Update README with usage instructions
2. Document configuration options
3. Provide Traefik integration examples
4. Add security best practices guide

## Success Metrics

- ✅ All authentication methods work correctly
- ✅ Route policies correctly allow/deny requests
- ✅ Header injection works with Traefik
- ✅ Configuration validation catches errors
- ✅ Health check endpoint responds
- ✅ Handles 1000+ req/s in benchmarks
- ✅ No secrets logged
- ✅ Graceful shutdown works

## Security Considerations

1. **Password/Token Comparison**: Use `crypto/subtle.ConstantTimeCompare` to prevent timing attacks
2. **JWT Validation**: Verify signature, issuer, audience, and expiration
3. **Logging**: Never log passwords, tokens, or API keys
4. **Headers**: Validate and sanitize all header values
5. **Configuration**: File permissions should be 0600 (readable only by owner)

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Traefik compatibility issues | Test with Traefik v2.x and v3.x |
| Performance bottlenecks | Use indexed maps, avoid regex where possible |
| Configuration errors | Provide validation command and clear error messages |
| Security vulnerabilities | Follow OWASP guidelines, use constant-time comparison |
| Streaming/SSE broken | Document that `forwardBody=true` must not be used |

## Alternatives Considered

1. **Use Traefik's built-in BasicAuth**
   - ❌ Only supports Basic Auth
   - ❌ Can't inject custom headers
   - ❌ No route-specific policies

2. **Use full OAuth2 provider (e.g., Keycloak)**
   - ❌ Too heavyweight for simple use cases
   - ❌ Requires database and complex setup
   - ✅ Would work, but overkill

3. **Custom middleware in each service**
   - ❌ Code duplication
   - ❌ Hard to maintain consistency
   - ❌ No centralized policy management

## Open Questions

1. **Hot Reload**: Should we implement configuration hot-reload in MVP? → **Decision: Post-MVP**, add file watcher later
2. **Metrics**: Include Prometheus metrics endpoint? → **Decision: Post-MVP**, focus on core functionality first
3. **Logging Format**: JSON or text logs? → **Decision: Both**, let user configure via env var

## References

- User-provided reference implementations
- [Traefik ForwardAuth Docs](https://doc.traefik.io/traefik/middlewares/http/forwardauth/)
- [Fiber Documentation](https://docs.gofiber.io/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
