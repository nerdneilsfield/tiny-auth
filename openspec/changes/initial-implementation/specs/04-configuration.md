# Spec: Configuration

## Overview

Define the TOML configuration structure and validation rules for tiny-auth.

## Configuration File Structure

```toml
# ===== Server Settings =====
[server]
port = "8080"
auth_path = "/auth"
health_path = "/health"
read_timeout = 5
write_timeout = 5

# ===== Header Configuration =====
[headers]
user_header = "X-Auth-User"
role_header = "X-Auth-Role"
method_header = "X-Auth-Method"
extra_headers = ["X-Auth-Timestamp"]
include_jwt_metadata = false

# ===== Basic Authentication =====
[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "supersecret"
roles = ["admin", "user"]

[[basic_auth]]
name = "dev-user"
user = "dev"
pass = "devpass"
roles = ["developer"]

# ===== Bearer Tokens =====
[[bearer_token]]
name = "prod-token"
token = "sk-live-abc123xyz789"
roles = ["admin", "service"]

[[bearer_token]]
name = "test-token"
token = "sk-test-def456uvw012"
roles = ["readonly"]

# ===== API Keys =====
[[api_key]]
name = "prod-key"
key = "ak_prod_xxx_secret"
roles = ["admin"]

[[api_key]]
name = "readonly-key"
key = "ak_read_yyy_secret"
roles = ["readonly"]

# ===== JWT Settings =====
[jwt]
secret = "your-256-bit-secret-key-here-must-be-at-least-32-chars-long"
issuer = "auth-service"
audience = "api"

# ===== Route Policies =====
[[route_policy]]
name = "public-api"
host = "api.public.com"
path_prefix = "/public"
allow_anonymous = true

[[route_policy]]
name = "admin-panel"
host = "admin.example.com"
allowed_basic_names = ["admin-user"]
require_all_roles = ["admin"]

[[route_policy]]
name = "webhook"
host = "hooks.example.com"
path_prefix = "/webhook"
allowed_bearer_names = ["prod-token"]
inject_authorization = "Bearer upstream-webhook-token"
```

## Configuration Sections

### [server]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `port` | string | No | `"8080"` | HTTP server port |
| `auth_path` | string | No | `"/auth"` | ForwardAuth endpoint path |
| `health_path` | string | No | `"/health"` | Health check endpoint path |
| `read_timeout` | int | No | `5` | Read timeout in seconds |
| `write_timeout` | int | No | `5` | Write timeout in seconds |

**Environment Variable Override:**
- `PORT` env var overrides `port` config

**Validation:**
- `port` must be valid port number (1-65535)
- `auth_path` and `health_path` must start with `/`
- Timeouts must be positive integers

### [headers]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `user_header` | string | No | `"X-Auth-User"` | Header name for username |
| `role_header` | string | No | `"X-Auth-Role"` | Header name for roles |
| `method_header` | string | No | `"X-Auth-Method"` | Header name for auth method |
| `extra_headers` | []string | No | `[]` | Additional headers to inject |
| `include_jwt_metadata` | bool | No | `false` | Include JWT claims as headers |

**Validation:**
- Header names must match: `^[A-Za-z][A-Za-z0-9-]*$`
- No duplicate names in `extra_headers`

### [[basic_auth]]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique identifier |
| `user` | string | Yes | - | Username |
| `pass` | string | Yes | - | Password (plaintext) |
| `roles` | []string | No | `["user"]` | Associated roles |

**Validation:**
- `name` must be unique across all `[[basic_auth]]` entries
- `user` must be non-empty
- `pass` must be non-empty
- `user` must be unique (no duplicate usernames)

**Security Note:**
- Store `config.toml` with permissions `0600` (owner read/write only)
- Consider using environment variables for passwords in production

### [[bearer_token]]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique identifier |
| `token` | string | Yes | - | Bearer token value |
| `roles` | []string | No | `["service"]` | Associated roles |

**Validation:**
- `name` must be unique across all `[[bearer_token]]` entries
- `token` must be non-empty
- `token` must be unique

### [[api_key]]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique identifier |
| `key` | string | Yes | - | API key value |
| `roles` | []string | No | `["api"]` | Associated roles |

**Validation:**
- `name` must be unique across all `[[api_key]]` entries
- `key` must be non-empty
- `key` must be unique

### [jwt]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `secret` | string | Yes* | - | HS256 signing secret |
| `issuer` | string | No | `""` | Expected issuer (iss claim) |
| `audience` | string | No | `""` | Expected audience (aud claim) |

**Note:** Entire `[jwt]` section is optional. If not present, JWT auth is disabled.

**Validation:**
- `secret` must be at least 32 characters (256 bits)
- `issuer` and `audience` are optional but recommended

### [[route_policy]]

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique identifier |
| `host` | string | No | `""` | Host pattern (exact or `*.domain`) |
| `path_prefix` | string | No | `""` | Path prefix to match |
| `method` | string | No | `""` | HTTP method (GET, POST, etc.) |
| `allow_anonymous` | bool | No | `false` | Allow unauthenticated access |
| `allowed_basic_names` | []string | No | `[]` | Allowed Basic Auth names |
| `allowed_bearer_names` | []string | No | `[]` | Allowed Bearer Token names |
| `allowed_api_key_names` | []string | No | `[]` | Allowed API Key names |
| `jwt_only` | bool | No | `false` | Only accept JWT |
| `require_all_roles` | []string | No | `[]` | User must have all roles |
| `require_any_role` | []string | No | `[]` | User must have any role |
| `inject_authorization` | string | No | `""` | Override Authorization header |

**Validation:**
- `name` must be unique across all `[[route_policy]]` entries
- Referenced names in `allowed_*` must exist in corresponding auth sections
- `allow_anonymous` conflicts with role requirements (warning)
- `jwt_only` conflicts with other `allowed_*` lists (warning)

## Configuration Loading

### Load Order

1. Read file from path (default: `config.toml`)
2. Parse TOML structure
3. Apply default values
4. Validate configuration
5. Build indexed data structures (AuthStore)
6. Log configuration summary

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `CONFIG_PATH` | Config file path | `/etc/tiny-auth/config.toml` |
| `PORT` | Override server port | `8080` |

### File Permissions Check

**On startup:**
- Check if `config.toml` permissions are too permissive
- Warn if readable by group/others: `chmod 0600 config.toml`
- Don't fail startup (just warn)

## Validation Command

```bash
tiny-auth validate [config-file]
```

**Checks:**
1. ✅ Valid TOML syntax
2. ✅ Required fields present
3. ✅ No duplicate names/users/tokens/keys
4. ✅ Valid header names
5. ✅ Valid policy references
6. ✅ JWT secret length (if present)
7. ✅ No conflicting policy settings
8. ⚠️  File permissions too permissive
9. ⚠️  Empty roles arrays
10. ⚠️  Weak passwords (if < 12 chars)

**Output:**
```
✓ Configuration is valid
✓ Server: port 8080, auth path /auth
✓ Basic Auth: 2 users configured
✓ Bearer Tokens: 2 tokens configured
✓ API Keys: 2 keys configured
✓ JWT: enabled with HS256
✓ Route Policies: 3 policies configured

⚠ Warning: config.toml has permissive permissions (0644)
  Recommendation: chmod 0600 config.toml

⚠ Warning: Basic auth 'dev-user' has short password (< 12 chars)
  Recommendation: Use longer passwords for security
```

## Example Configurations

### Minimal Configuration

```toml
[server]
port = "8080"

[[basic_auth]]
name = "user"
user = "admin"
pass = "secret"
roles = ["user"]
```

### Production Configuration

```toml
[server]
port = "8080"
auth_path = "/auth"
health_path = "/health"
read_timeout = 10
write_timeout = 10

[headers]
user_header = "X-Auth-User"
role_header = "X-Auth-Role"
method_header = "X-Auth-Method"
extra_headers = ["X-Auth-Timestamp"]

# Production users
[[basic_auth]]
name = "admin"
user = "admin"
pass = "${ADMIN_PASSWORD}"  # Note: env var substitution not implemented yet
roles = ["admin", "user"]

# Production API tokens
[[bearer_token]]
name = "api-prod"
token = "${API_TOKEN}"
roles = ["api", "service"]

# JWT validation
[jwt]
secret = "${JWT_SECRET}"
issuer = "prod-auth-service"
audience = "prod-api"

# Admin panel requires admin role
[[route_policy]]
name = "admin-panel"
host = "admin.example.com"
require_all_roles = ["admin"]

# Public endpoints
[[route_policy]]
name = "public"
path_prefix = "/public"
allow_anonymous = true
```

**Note:** Environment variable substitution (`${VAR}`) is a future enhancement.

### Development Configuration

```toml
[server]
port = "8080"

[[basic_auth]]
name = "dev"
user = "dev"
pass = "dev"
roles = ["developer"]

[[basic_auth]]
name = "admin"
user = "admin"
pass = "admin"
roles = ["admin", "developer"]

# Allow all routes
# (no route policies = default to authenticated access)
```

## Configuration Hot-Reload (Future)

**Not in MVP**, but planned:

```bash
# Send SIGHUP to reload config
kill -HUP $(pidof tiny-auth)

# Or use HTTP endpoint
curl -X POST http://localhost:8080/reload
```

**Behavior:**
- Reload config file
- Rebuild AuthStore
- Log config changes
- No downtime (atomic swap)

## Test Scenarios

### Scenario 1: Valid Minimal Config
```toml
[server]
port = "8080"

[[basic_auth]]
name = "user1"
user = "admin"
pass = "secret"
```

**Expected:** ✅ Valid configuration

### Scenario 2: Duplicate Username
```toml
[[basic_auth]]
name = "user1"
user = "admin"
pass = "secret1"

[[basic_auth]]
name = "user2"
user = "admin"  # ❌ Duplicate
pass = "secret2"
```

**Expected:** ❌ Validation error: "duplicate username 'admin'"

### Scenario 3: Invalid Policy Reference
```toml
[[route_policy]]
name = "policy1"
allowed_basic_names = ["nonexistent"]  # ❌ Does not exist
```

**Expected:** ❌ Validation error: "policy 'policy1' references unknown basic auth 'nonexistent'"

### Scenario 4: Conflicting Anonymous + Roles
```toml
[[route_policy]]
name = "conflict"
allow_anonymous = true
require_all_roles = ["admin"]  # ⚠️  Conflict
```

**Expected:** ⚠️  Warning: "policy 'conflict' allows anonymous but requires roles (roles ignored)"

### Scenario 5: Short JWT Secret
```toml
[jwt]
secret = "short"  # ❌ Less than 32 chars
```

**Expected:** ❌ Validation error: "JWT secret must be at least 32 characters"

### Scenario 6: Invalid Header Name
```toml
[headers]
user_header = "X Auth User"  # ❌ Contains space
```

**Expected:** ❌ Validation error: "invalid header name 'X Auth User'"

### Scenario 7: Config File Not Found
```bash
tiny-auth validate missing.toml
```

**Expected:** ❌ Error: "failed to load config: open missing.toml: no such file or directory"
