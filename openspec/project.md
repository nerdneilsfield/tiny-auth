# tiny-auth

A lightweight, high-performance authentication service for Traefik ForwardAuth middleware.

## Overview

tiny-auth is a Go-based authentication service designed to work seamlessly with Traefik's ForwardAuth middleware. It provides multiple authentication methods (Basic Auth, Bearer Token, API Key, JWT) with flexible route-based policy control and header injection capabilities.

## Purpose

### Problem Statement

When using Traefik as a reverse proxy, services often need:
- Centralized authentication before requests reach upstream services
- Multiple authentication methods for different use cases
- Route-specific authentication policies (different auth for different paths/hosts)
- Ability to inject authentication headers to upstream services
- Support for services that require specific auth formats (e.g., client uses Basic Auth, upstream needs Bearer)

### Solution

tiny-auth acts as a ForwardAuth authentication service that:
1. Receives authentication requests from Traefik
2. Validates credentials against configured policies
3. Returns appropriate HTTP status (200 for success, 401/403 for failure)
4. Injects configured headers into upstream requests (via Traefik's `authResponseHeaders`)

## Core Features

### Authentication Methods

1. **Basic Auth**
   - Multiple user/password combinations
   - Per-user role assignment
   - Constant-time password comparison

2. **Bearer Token**
   - Multiple token configurations
   - Per-token role assignment
   - Support for both JWT and static tokens

3. **API Key**
   - Support via `Authorization: ApiKey xxx` header
   - Support via `X-Api-Key` header
   - Multiple API key configurations with roles

4. **JWT**
   - HS256 signature validation
   - Issuer and audience verification
   - Custom claims support

### Route Policies

Route-based authentication policies support:
- Host matching (exact and wildcard: `*.example.com`)
- Path prefix matching
- HTTP method matching
- Allow anonymous access for specific routes
- Restrict to specific authentication methods
- Require specific roles (all or any)

### Header Injection

- Inject custom headers to upstream requests
- Support for user/role headers (`X-Auth-User`, `X-Auth-Role`)
- Support for method header (`X-Auth-Method`)
- Support for custom headers (e.g., `X-Auth-Timestamp`)
- Override `Authorization` header for upstream services

## Architecture

### Components

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Client    │────────▶│   Traefik   │────────▶│  tiny-auth  │
└─────────────┘         └─────────────┘         └─────────────┘
                               │                        │
                               │  (on success)          │
                               ▼                        ▼
                        ┌─────────────┐         ┌─────────────┐
                        │  Upstream   │◀────────│  Returns    │
                        │  Service    │         │  200 + HDR  │
                        └─────────────┘         └─────────────┘
```

### Directory Structure

```
tiny-auth/
├── cmd/
│   ├── root.go           # CLI root command
│   ├── server.go         # Server start command
│   ├── validate.go       # Config validation command
│   └── version.go        # Version command
├── internal/
│   ├── config/
│   │   ├── config.go     # Configuration structures
│   │   └── loader.go     # TOML config loader
│   ├── auth/
│   │   ├── basic.go      # Basic Auth handler
│   │   ├── bearer.go     # Bearer Token handler
│   │   ├── apikey.go     # API Key handler
│   │   ├── jwt.go        # JWT handler
│   │   └── store.go      # Auth store (indexed config)
│   ├── policy/
│   │   ├── matcher.go    # Route policy matching
│   │   └── checker.go    # Policy enforcement
│   └── server/
│       ├── server.go     # Fiber HTTP server
│       └── handler.go    # ForwardAuth handler
├── pkg/
│   └── middleware/
│       └── logger.go     # Request logging middleware
├── config.toml           # Example configuration
├── main.go               # Entry point
└── README.md
```

## Configuration

### TOML Structure

Configuration uses TOML format with the following sections:

1. **[server]** - Server settings (port, paths)
2. **[headers]** - Header configuration for injection
3. **[[basic_auth]]** - Basic auth credentials (multiple)
4. **[[bearer_token]]** - Bearer tokens (multiple)
5. **[[api_key]]** - API keys (multiple)
6. **[jwt]** - JWT validation settings
7. **[[route_policy]]** - Route-specific policies (multiple)

### Example Flow

```toml
[server]
port = "8080"
auth_path = "/auth"

[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "secret"
roles = ["admin"]

[[route_policy]]
name = "admin-panel"
host = "admin.example.com"
allowed_basic_names = ["admin-user"]
require_all_roles = ["admin"]
```

## Integration with Traefik

### Docker Labels Example

```yaml
services:
  tiny-auth:
    image: tiny-auth:latest
    environment:
      - CONFIG_PATH=/config/config.toml
    volumes:
      - ./config.toml:/config/config.toml

  app:
    image: your-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.app.rule=Host(`app.example.com`)"
      
      # ForwardAuth middleware
      - "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
      - "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=Authorization,X-Auth-User,X-Auth-Role"
      
      - "traefik.http.routers.app.middlewares=auth@docker"
```

### Key Traefik Settings

- `address`: tiny-auth's `/auth` endpoint
- `authResponseHeaders`: Headers to copy from auth response to upstream request
- `trustForwardHeader`: Whether to trust existing `X-Forwarded-*` headers
- `forwardBody`: **Never enable** (breaks streaming/SSE)

## Non-Goals

- **User Management UI**: This is a configuration-file-based service, not a user management system
- **Database Integration**: All configuration is file-based (TOML)
- **OAuth/OIDC Provider**: Only validates JWT tokens, doesn't act as an identity provider
- **Rate Limiting**: Use Traefik's built-in rate limiting middleware
- **Session Management**: Stateless authentication only

## Success Criteria

1. **Performance**: Handle 1000+ req/s on modest hardware
2. **Security**: 
   - Constant-time comparison for passwords/tokens
   - No secrets in logs
   - Secure default headers
3. **Compatibility**: Works with Traefik v2.x and v3.x
4. **Usability**: 
   - Clear error messages
   - Configuration validation command
   - Example configurations in docs
5. **Reliability**:
   - Graceful shutdown
   - Health check endpoint
   - Structured logging

## Tech Stack

- **Language**: Go 1.23+
- **HTTP Framework**: Fiber v2 (high-performance, Express-like API)
- **Config Format**: TOML (via BurntSushi/toml)
- **JWT**: golang-jwt/jwt/v5
- **CLI**: spf13/cobra
- **Logging**: uber-go/zap

## Development Workflow

1. Define/update specs in `openspec/`
2. Implement features following Go best practices
3. Write tests for each authentication method
4. Validate configuration examples
5. Update documentation
6. Run integration tests with Traefik

## Future Enhancements (Post-MVP)

- Configuration hot-reload (file watch)
- Prometheus metrics endpoint
- Redis-based distributed rate limiting
- LDAP/Active Directory integration
- mTLS client certificate authentication
- OpenTelemetry tracing support

## References

- [Traefik ForwardAuth Documentation](https://doc.traefik.io/traefik/middlewares/http/forwardauth/)
- [Fiber Framework](https://docs.gofiber.io/)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- Reference implementations provided by user (see change proposals)
