# Example Configurations

[简体中文](README_ZH.md) | English

This directory contains configuration examples for various use cases.

## File Descriptions

- **docker-compose-full.yml** - Complete Traefik + tiny-auth example
  - Includes 5 different authentication scenarios
  - Demonstrates multiple authentication methods
  - Ready-to-use complete environment

- **config-full.toml** - Corresponding complete configuration file
  - Contains all authentication method configurations
  - Demonstrates route policy usage
  - Uses environment variables for sensitive information

## Quick Start

### 1. Prepare Configuration File

```bash
cd examples
cp config-full.toml config.toml
```

### 2. Set Environment Variables (Optional)

Create `.env` file:

```bash
cat > .env << 'EOF'
ADMIN_PASSWORD=your-secure-admin-password
DEV_PASSWORD=your-dev-password
API_TOKEN=your-api-token-here
JWT_SECRET=your-jwt-secret-key-must-be-at-least-32-chars
EOF
```

### 3. Start Services

```bash
docker-compose -f docker-compose-full.yml up -d
```

### 4. Test Various Scenarios

#### Scenario 1: Basic Auth

```bash
# Success - admin user
curl -u admin:your-secure-admin-password http://whoami-basic.localhost/

# Success - dev user
curl -u dev:your-dev-password http://whoami-basic.localhost/

# Failure - wrong password
curl -u admin:wrong http://whoami-basic.localhost/
# → 401 Unauthorized
```

#### Scenario 2: Bearer Token

```bash
# Success
curl -H "Authorization: Bearer your-api-token-here" http://api.localhost/

# Failure - invalid token
curl -H "Authorization: Bearer invalid" http://api.localhost/
# → 401 Unauthorized
```

#### Scenario 3: Public Access (No Auth)

```bash
# Success - anonymous access
curl http://public.localhost/public/
```

#### Scenario 4: Admin Panel (Admin Only)

```bash
# Success - admin user
curl -u admin:your-secure-admin-password http://admin.localhost/

# Failure - dev user (no admin role)
curl -u dev:your-dev-password http://admin.localhost/
# → 401 Unauthorized (policy requirements not met)
```

#### Scenario 5: API Key Authentication

```bash
# Success - using X-Api-Key header
curl -H "X-Api-Key: ak_internal_secret_key" http://internal.localhost/

# Success - using Authorization header
curl -H "Authorization: ApiKey ak_internal_secret_key" http://internal.localhost/
```

### 5. View Authentication Headers

All successful requests receive injected authentication headers:

```bash
curl -v -u admin:your-secure-admin-password http://whoami-basic.localhost/

# Response headers include:
# X-Auth-User: admin
# X-Auth-Role: admin,user
# X-Auth-Method: basic
# X-Auth-Timestamp: 1738560000
```

### 6. Health Check

```bash
# tiny-auth health check
curl http://auth.localhost/health

{
  "status": "ok",
  "basic_count": 2,
  "bearer_count": 1,
  "apikey_count": 1,
  "jwt_enabled": true,
  "policy_count": 5
}
```

### 7. Traefik Dashboard

Visit http://traefik.localhost:8080 to view Traefik Dashboard.

## Configuration Details

### Authentication Flow

```
Client → Traefik (detects auth needed) → tiny-auth (/auth)
                                         ↓
                                    Verify credentials
                                         ↓
                                    Check policy
                                         ↓
                          Return 200 + Headers / 401
                                         ↓
          Traefik (inject Headers) → Upstream service
```

### ForwardAuth Configuration Points

```yaml
labels:
  # ForwardAuth address
  - "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
  
  # Headers to inject to upstream
  - "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
  
  # Don't trust existing X-Forwarded-* headers (security)
  - "traefik.http.middlewares.auth.forwardauth.trustForwardHeader=false"
  
  # Apply middleware to route
  - "traefik.http.routers.myservice.middlewares=auth@docker"
```

⚠️ **Important**:
- Don't use `forwardBody=true`, it breaks SSE/WebSocket
- Ensure `authResponseHeaders` includes all headers you need
- Recommend setting `trustForwardHeader` to `false`

## Stop Services

```bash
docker-compose -f docker-compose-full.yml down
```

## Troubleshooting

### Issue: Always Returns 401

1. Check tiny-auth logs
   ```bash
   docker logs tiny-auth
   ```

2. Check configuration file
   ```bash
   docker exec tiny-auth cat /root/config.toml
   ```

3. Validate configuration
   ```bash
   docker exec tiny-auth ./tiny-auth validate /root/config.toml
   ```

### Issue: Headers Not Passed to Upstream

Ensure Traefik's `authResponseHeaders` includes required headers:

```yaml
- "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
```

### Issue: Environment Variables Not Working

1. Check if `.env` file is in correct location
2. Ensure config uses `env:VAR_NAME` syntax
3. Restart services for environment variables to take effect

## More Examples

Check the `docs/` directory in the main repository for more configuration examples and best practices.
