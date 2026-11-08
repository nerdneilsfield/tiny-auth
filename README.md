# 🔐 tiny-auth

<div align="center">

**The Most Lightweight Traefik ForwardAuth Service Ever! 🚀**

_One config file to rule them all, never worry about API security again!_

[![Go Version](https://img.shields.io/github/go-mod/go-version/nerdneilsfield/tiny-auth?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/nerdneilsfield/tiny-auth?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/nerdneilsfield/tiny-auth?style=flat-square&logo=github)](https://github.com/nerdneilsfield/tiny-auth/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/nerdneils/tiny-auth?style=flat-square&logo=docker)](https://hub.docker.com/r/nerdneils/tiny-auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/nerdneilsfield/tiny-auth?style=flat-square)](https://goreportcard.com/report/github.com/nerdneilsfield/tiny-auth)
[![Build Status](https://img.shields.io/github/actions/workflow/status/nerdneilsfield/tiny-auth/goreleaser.yml?style=flat-square&logo=github-actions)](https://github.com/nerdneilsfield/tiny-auth/actions)

English | [简体中文](README_ZH.md)

</div>

---

## ✨ Why tiny-auth?

> 💡 **TL;DR**: If you're using Traefik as a reverse proxy and struggling with authentication, tiny-auth is your lifesaver!

### 🎯 Key Features

- 🪶 **Ultra Lightweight**: Single binary, zero dependencies, under 5MB
- ⚡ **Blazingly Fast**: Powered by Fiber framework, easily handles 1000+ req/s
- 🔒 **Security First**: Constant-time comparison prevents timing attacks, config file permission checks
- 🎨 **Flexible Config**: TOML format, environment variable support, hot reload without restart
- 🌈 **Multiple Auth Methods**: Basic Auth / Bearer Token / API Key / JWT all-in-one
- 🎯 **Fine-grained Control**: Route-based policies by host/path/method, play it your way
- 🔄 **Header Injection**: Client uses Basic Auth, upstream receives Bearer Token? No problem!
- 📊 **Out-of-the-box**: Health checks, debug endpoints, config validation, everything you need

### 🆚 Comparison

| Feature | tiny-auth | Traefik BasicAuth | OAuth2 Proxy | Authelia |
|---------|-----------|-------------------|--------------|----------|
| Binary Size | ~5MB | N/A | ~20MB | ~30MB |
| Multiple Auth | ✅ | ❌ | ✅ | ✅ |
| Route Policies | ✅ | ❌ | ❌ | ✅ |
| Header Transform | ✅ | ❌ | ⚠️ | ❌ |
| Hot Reload | ✅ | ❌ | ❌ | ✅ |
| Env Vars | ✅ | ✅ | ✅ | ✅ |
| Complexity | ⭐ | ⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |

---

## 📦 Docker Images

tiny-auth is available from two registries for fast access worldwide:

| Registry | Address | Recommended For |
|----------|---------|-----------------|
| 🐳 **Docker Hub** | `nerdneils/tiny-auth:latest` | 🌍 International |
| 📦 **GitHub CR** | `ghcr.io/nerdneilsfield/tiny-auth:latest` | 🌏 Asia (may be slower) |

**Supported Architectures**:
- `linux/amd64` - x86_64
- `linux/arm64` - ARM64 / Apple Silicon
- `linux/arm/v7` - ARMv7

## Homebrew / Scoop

```bash
# macOS / Linux
brew install --cask nerdneilsfield/apps/tiny-auth
```

```powershell
# Windows
scoop bucket add nerdneilsfield https://github.com/nerdneilsfield/homebrew-apps
scoop install nerdneilsfield/tiny-auth
```

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# 1. Create config file
cat > config.toml << 'EOF'
[server]
port = "8080"

[[basic_auth]]
name = "admin"
user = "admin"
pass = "supersecret"
roles = ["admin"]
EOF

# 2. Run (Choose one registry)
# Docker Hub
docker run -d \
  --name tiny-auth \
  -p 8080:8080 \
  -v $(pwd)/config.toml:/root/config.toml:ro \
  nerdneils/tiny-auth:latest

# Or use GitHub Container Registry
docker run -d \
  --name tiny-auth \
  -p 8080:8080 \
  -v $(pwd)/config.toml:/root/config.toml:ro \
  ghcr.io/nerdneilsfield/tiny-auth:latest

# 3. Test
curl -u admin:supersecret http://localhost:8080/auth
# → 200 OK ✅
```

### Option 2: Binary Download

```bash
# Download latest release
wget https://github.com/nerdneilsfield/tiny-auth/releases/latest/download/tiny-auth_linux_amd64.tar.gz
tar -xzf tiny-auth_linux_amd64.tar.gz

# Run
./tiny-auth server --config config.toml
```

### Option 3: Build from Source

```bash
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth
just build  # or make build
./tiny-auth server
```

---

## 🎨 Configuration Examples

<details>
<summary><b>📖 Complete Configuration Example (Click to expand)</b></summary>

```toml
# ===== Server Configuration =====
[server]
port = "8080"
auth_path = "/auth"
health_path = "/health"

# ===== Logging Configuration =====
[logging]
format = "json"  # or "text"
level = "info"   # debug/info/warn/error

# ===== Basic Auth =====
[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "supersecret"        # Supports env:PASSWORD to read from environment
roles = ["admin", "user"]

# ===== Bearer Token =====
[[bearer_token]]
name = "api-token"
token = "env:API_TOKEN"     # Read from environment variable
roles = ["api", "service"]

# ===== API Key =====
[[api_key]]
name = "prod-key"
key = "ak_prod_xxx"
roles = ["admin"]

# ===== JWT Validation =====
[jwt]
secret = "your-256-bit-secret-key-must-be-32-chars"
issuer = "auth-service"
audience = "api"

# ===== Route Policies =====
# Public API allows anonymous access
[[route_policy]]
name = "public"
path_prefix = "/public"
allow_anonymous = true

# Admin panel requires admin role
[[route_policy]]
name = "admin"
host = "admin.example.com"
allowed_basic_names = ["admin-user"]
require_all_roles = ["admin"]

# Webhook endpoint with header injection
[[route_policy]]
name = "webhook"
host = "hooks.example.com"
path_prefix = "/webhook"
allowed_bearer_names = ["api-token"]
inject_authorization = "Bearer upstream-token-123"
```

</details>

<details>
<summary><b>🔑 Environment Variable Syntax</b></summary>

Use `env:VAR_NAME` in config to read sensitive data from environment:

```toml
[[basic_auth]]
pass = "env:ADMIN_PASSWORD"

[jwt]
secret = "env:JWT_SECRET"
```

Set environment variables when starting:

```bash
export ADMIN_PASSWORD="my-secure-password"
export JWT_SECRET="my-jwt-secret-key-32-chars-long"
./tiny-auth server
```

</details>

---

## 🔌 Traefik Integration

### Complete Docker Compose Example

```yaml
version: '3.8'

services:
  # tiny-auth authentication service
  tiny-auth:
    image: nerdneils/tiny-auth:latest  # Or use ghcr.io/nerdneilsfield/tiny-auth:latest
    volumes:
      - ./config.toml:/root/config.toml:ro
    networks:
      - traefik

  # Traefik reverse proxy
  traefik:
    image: traefik:v3.2
    ports:
      - "80:80"
    command:
      - --providers.docker=true
      - --entrypoints.web.address=:80
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - traefik

  # Protected service (example)
  whoami:
    image: traefik/whoami
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.whoami.rule=Host(`whoami.localhost`)"
      
      # Configure ForwardAuth middleware
      - "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
      - "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
      
      # Apply middleware
      - "traefik.http.routers.whoami.middlewares=auth@docker"
    networks:
      - traefik

networks:
  traefik:
```

### Key Configuration Settings

| Setting | Description | Example |
|---------|-------------|---------|
| `address` | tiny-auth's /auth endpoint | `http://tiny-auth:8080/auth` |
| `authResponseHeaders` | Headers to inject to upstream | `X-Auth-User,X-Auth-Role` |
| `trustForwardHeader` | Trust X-Forwarded-* headers | `false` (recommended) |

⚠️ **Important**: Never enable `forwardBody=true`, it breaks SSE/WebSocket!

---

## 🎯 Use Cases

### Case 1: Multi-environment API Authentication

```toml
# Development environment uses Basic Auth
[[basic_auth]]
name = "dev"
user = "dev"
pass = "dev123"
roles = ["developer"]

# Production environment uses Bearer Token
[[bearer_token]]
name = "prod"
token = "env:PROD_TOKEN"
roles = ["admin", "service"]

# Route policies
[[route_policy]]
name = "dev-api"
host = "dev-api.example.com"
allowed_basic_names = ["dev"]

[[route_policy]]
name = "prod-api"
host = "api.example.com"
allowed_bearer_names = ["prod"]
```

### Case 2: Auth Transformation (Client Basic → Upstream Bearer)

```toml
[[basic_auth]]
name = "user"
user = "client"
pass = "clientpass"
roles = ["user"]

[[route_policy]]
name = "transform"
host = "api.example.com"
allowed_basic_names = ["user"]
inject_authorization = "Bearer upstream-internal-token-abc123"
```

Client uses Basic Auth, upstream service receives Bearer Token!

### Case 3: Microservices Internal Authentication

```toml
# Service-to-service communication with API Keys
[[api_key]]
name = "service-a"
key = "env:SERVICE_A_KEY"
roles = ["internal"]

[[api_key]]
name = "service-b"
key = "env:SERVICE_B_KEY"
roles = ["internal"]

[[route_policy]]
name = "internal"
host = "internal.example.com"
require_any_role = ["internal"]
```

---

## 🐳 Docker Usage Guide

<details>
<summary><b>🏃 Quick Start Full Environment</b></summary>

Use our complete example (includes Traefik + tiny-auth + 5 demo services):

```bash
# 1. Clone repository
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth/examples

# 2. Prepare configuration
cp config-full.toml config.toml
cp .env.example .env
# Edit .env file to set your passwords

# 3. Start services
docker-compose -f docker-compose-full.yml up -d

# 4. Test
curl -u admin:your-password http://whoami-basic.localhost/
curl http://public.localhost/public/
curl -H "Authorization: Bearer your-token" http://api.localhost/

# 5. View logs
docker-compose -f docker-compose-full.yml logs -f tiny-auth

# 6. Stop services
docker-compose -f docker-compose-full.yml down
```

</details>

<details>
<summary><b>🔧 Custom Docker Compose</b></summary>

Integrate tiny-auth in your project:

```yaml
version: '3.8'

services:
  tiny-auth:
    image: nerdneils/tiny-auth:latest
    volumes:
      - ./your-config.toml:/root/config.toml:ro
    environment:
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
    networks:
      - traefik

  your-service:
    image: your-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.app.rule=Host(`app.example.com`)"
      - "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
      - "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role"
      - "traefik.http.routers.app.middlewares=auth@docker"
    networks:
      - traefik
```

</details>

<details>
<summary><b>🔄 Hot Reload Configuration</b></summary>

After modifying config file, no container restart needed:

```bash
# Method 1: Send SIGHUP signal
docker kill --signal=SIGHUP tiny-auth

# Method 2: Use docker-compose
docker-compose kill -s SIGHUP tiny-auth

# View reload logs
docker logs tiny-auth --tail 20
# → "♻️  Configuration reloaded"
```

</details>

<details>
<summary><b>🏗️ Build from Source</b></summary>

```bash
# Clone repository
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth

# Build image
docker build -t my-tiny-auth:latest .

# Or use justfile
just docker-build

# Run
docker run -d -p 8080:8080 \
  -v $(pwd)/config.toml:/root/config.toml:ro \
  my-tiny-auth:latest
```

</details>

---

## 📊 Operations

### Hot Reload Configuration

```bash
# Send SIGHUP signal to reload config
kill -HUP $(pidof tiny-auth)

# Or with Docker
docker kill --signal=SIGHUP tiny-auth
```

No restart needed, config takes effect immediately! ✨

### Configuration Validation

```bash
# Validate config file
tiny-auth validate config.toml

# Output example
✅ Configuration is valid

📋 Configuration Summary:
✓ Server: port 8080
✓ Basic Auth: 2 users configured
✓ Bearer Tokens: 1 tokens configured
✓ Route Policies: 3 policies configured

⚠ Warning: config.toml has insecure permissions 644
⚠ Recommendation: chmod 0600 config.toml
```

### Health Check

```bash
curl http://localhost:8080/health

{
  "status": "ok",
  "basic_count": 2,
  "bearer_count": 1,
  "apikey_count": 1,
  "jwt_enabled": true,
  "policy_count": 3
}
```

### Debug Endpoint (Optional)

Enable in config first:

```toml
[server]
enable_debug = true
```

```bash
curl http://localhost:8080/debug/config

{
  "server": {
    "port": "8080",
    "auth_path": "/auth"
  },
  "authentication": {
    "basic_auth": ["admin", "dev"],
    "bearer_tokens": ["prod-token"],
    "jwt_enabled": true
  },
  "policies": ["public", "admin-only"]
}
```

⚠️ **Do not expose this endpoint publicly.** Restrict it to trusted networks only.

### Audit Logging (Recommended for Production)

tiny-auth can emit a **separate structured audit log stream** (JSON Lines) for compliance and traceability.

```toml
[audit]
enabled = true
output = "./audit.log"   # or "stdout" / "stderr"
```

Each line is a JSON event including request_id, client_ip, host/uri, auth_method, policy, result, status, and latency.

---

## 🔒 Security Best Practices

### ✅ Must Do

1. **⚠️ Configure Trusted Proxies (CRITICAL!)**
   
   **Why**: Prevents attackers from spoofing `X-Forwarded-*` headers to bypass policies.

   **Important**: Your reverse proxy/load balancer MUST strip or overwrite any client-supplied `X-Forwarded-*` headers.  
   If it doesn't, `trusted_proxies` can still be bypassed.
   
   ```toml
   [server]
   # ✅ PRODUCTION: Only trust your reverse proxy
   trusted_proxies = ["172.16.0.0/12"]  # Docker network
   
   # ❌ INSECURE: Empty list accepts headers from ANY source
   # trusted_proxies = []
   ```
   
   **Examples**:
   - Docker Compose: `["172.16.0.0/12"]`
   - Kubernetes: `["10.0.0.0/8"]`
   - Specific IP: `["192.168.1.100"]`
   - Multiple: `["172.16.0.0/12", "192.168.1.0/24"]`
   
   **Without it**:
   ```bash
   # Attacker can fake host to bypass policies
   curl -H "X-Forwarded-Host: admin.internal.com" \
        http://your-tiny-auth:8080/auth
   # Without trusted_proxies: ✅ Allowed (bypass!)
   # With trusted_proxies:    ❌ Denied (headers ignored)
   ```

2. **Config File Permissions**
   ```bash
   chmod 0600 config.toml  # Owner read/write only
   ```

3. **Use Environment Variables for Secrets**
   ```toml
   pass = "env:ADMIN_PASSWORD"  # ✅
   pass = "plaintext123"        # ❌
   ```

4. **Strong Password Policy**
   - At least 12 characters
   - Mix of uppercase, lowercase, numbers, special chars

5. **JWT Secret Length**
   - At least 32 characters (256 bits)

6. **Enable JSON Logging for Production**
   ```toml
   [logging]
   format = "json"  # Structured, searchable
   level = "info"
   ```
   
   **Structured logs include**:
   - `request_id`: Trace across services
   - `client_ip`: Real client IP (validated via trusted_proxies)
   - `auth_method`: Which authentication succeeded
   - `latency`: Performance monitoring
   - `reason`: Why authentication failed

### ⚠️ Warnings

- Don't expose `/debug/config` endpoint publicly
- Rotate tokens and API keys regularly
- Use HTTPS (configure TLS in Traefik)
- Review auth logs, monitor suspicious access

---

## 🛠️ Development Guide

<details>
<summary><b>🔨 Local Development</b></summary>

### Prerequisites

- Go 1.23+
- [just](https://github.com/casey/just) or make

### Common Commands

```bash
# Clone repository
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth

# Install dependencies
just deps

# Build
just build

# Run tests
just test

# Lint
just lint

# Format code
just fmt

# Full check (test + lint)
just check

# List all commands
just --list
```

### Project Structure

```
tiny-auth/
├── cmd/                # CLI commands
│   ├── root.go        # Root command
│   ├── server.go      # Server command
│   ├── validate.go    # Config validation
│   └── version.go     # Version info
├── internal/          # Internal packages
│   ├── config/        # Config management
│   ├── auth/          # Authentication logic
│   ├── policy/        # Policy matching
│   └── server/        # HTTP server
├── openspec/          # OpenSpec specification
└── main.go            # Entry point
```

</details>

<details>
<summary><b>🧪 Testing</b></summary>

```bash
# Run all tests
just test

# Generate coverage report
just test-coverage
open coverage.html

# Race detection
go test -race ./...
```

Current coverage target: >80%

</details>

---

## 📚 Documentation

- [Complete Configuration Reference](openspec/changes/initial-implementation/specs/04-configuration.md)
- [Authentication Methods](openspec/changes/initial-implementation/specs/01-authentication.md)
- [Route Policies](openspec/changes/initial-implementation/specs/02-route-policies.md)
- [Header Injection](openspec/changes/initial-implementation/specs/03-header-injection.md)
- [Technical Design](openspec/changes/initial-implementation/design.md)

---

## 🤝 Contributing

All contributions are welcome!

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Before committing:

```bash
just pre-commit  # Format + check
```

---

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgments

- [Fiber](https://github.com/gofiber/fiber) - High-performance web framework
- [Traefik](https://github.com/traefik/traefik) - Modern reverse proxy
- [golang-jwt](https://github.com/golang-jwt/jwt) - JWT implementation
- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML parser

---

## 💬 Community & Support

- 🐛 [Report Issues](https://github.com/nerdneilsfield/tiny-auth/issues)
- 💡 [Feature Requests](https://github.com/nerdneilsfield/tiny-auth/discussions)
- 📧 Contact: dengqi935@gmail.com

---

<div align="center">

**⭐ Star this repo if you find it useful! ⭐**

Made with ❤️ by [dengqi](https://github.com/nerdneilsfield)

</div>
