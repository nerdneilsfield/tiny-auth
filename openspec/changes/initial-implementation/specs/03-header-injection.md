# Spec: Header Injection

## Overview

Define how tiny-auth injects headers into authenticated requests for upstream services.

## Requirements

### REQ-HEADER-001: Response Headers

**Given** authentication succeeds  
**When** returning 200 OK to Traefik  
**Then** include configured headers in HTTP response

**Mechanism:**
- Traefik's `authResponseHeaders` configuration specifies which headers to copy
- tiny-auth sets headers in its 200 response
- Traefik copies those headers to the upstream request

**Example Traefik Config:**
```yaml
- "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=Authorization,X-Auth-User,X-Auth-Role"
```

### REQ-HEADER-002: Standard Headers

**Configuration:**
```toml
[headers]
user_header = "X-Auth-User"      # Username or subject
role_header = "X-Auth-Role"      # Comma-separated roles
method_header = "X-Auth-Method"  # Auth method used
```

**Behavior:**

1. **X-Auth-Method** (always set)
   - Value: `"basic"`, `"bearer"`, `"apikey"`, `"jwt"`, or `"anonymous"`

2. **X-Auth-User** (if available)
   - Basic Auth: username (e.g., `"admin"`)
   - JWT: subject claim (e.g., `"user123"`)
   - Bearer/API Key: credential name if no explicit user
   - Anonymous: not set

3. **X-Auth-Role** (if available)
   - Comma-separated list: `"admin,user"`
   - Empty if no roles

**Example:**
```
X-Auth-Method: basic
X-Auth-User: admin
X-Auth-Role: admin,user
```

### REQ-HEADER-003: Custom Headers

**Configuration:**
```toml
[headers]
extra_headers = ["X-Auth-Timestamp", "X-Auth-Route"]
```

**Behavior:**
- `X-Auth-Timestamp`: Unix timestamp (seconds)
- `X-Auth-Route`: Original host + path (e.g., `api.example.com/api/users`)
- Custom logic per header name

**Example:**
```
X-Auth-Timestamp: 1738560000
X-Auth-Route: api.example.com/api/users
```

### REQ-HEADER-004: Authorization Header Transformation

**Use Case:** Client authenticates with Basic, but upstream needs Bearer token.

**Configuration:**
```toml
[[route_policy]]
name = "transform-auth"
host = "api.example.com"
inject_authorization = "Bearer upstream-token-abc123"
```

**Behavior:**
- When this policy matches AND auth succeeds
- Set `Authorization: Bearer upstream-token-abc123` in response
- Traefik replaces client's Authorization header with this value
- Upstream receives transformed header

**Example Flow:**
```
Client → Traefik: Authorization: Basic YWRtaW46c2VjcmV0
Traefik → tiny-auth: Authorization: Basic YWRtaW46c2VjcmV0
tiny-auth → Traefik: 200 OK, Authorization: Bearer upstream-token
Traefik → Upstream: Authorization: Bearer upstream-token
```

**Note:** Requires `Authorization` in `authResponseHeaders` to take effect.

### REQ-HEADER-005: Metadata Headers

**For JWT authentication**, include JWT claims as headers:

**Example:**
- JWT `iss` claim → `X-Auth-Issuer: auth-service`
- JWT `aud` claim → `X-Auth-Audience: api`
- JWT `exp` claim → `X-Auth-Expires: 1738560000`

**Prefix:** All metadata headers use `X-Auth-` prefix.

**Configuration:**
```toml
[headers]
include_jwt_metadata = true  # Default: false
```

## Header Precedence

If multiple headers would set the same name:

1. **Policy-specific** `inject_authorization` overrides all
2. **Authentication result** (user/role/method) takes precedence
3. **Custom/metadata** headers last

## Security Considerations

### REQ-HEADER-SEC-001: No Sensitive Data

**Must not** include in headers:
- Passwords
- Token values
- API keys
- JWT signatures

**May include:**
- Usernames
- Roles
- Token/key names
- Timestamps

### REQ-HEADER-SEC-002: Header Validation

**Must sanitize** all header values:
- Remove newlines (`\r`, `\n`)
- Limit length (max 1024 bytes per header)
- Remove non-printable characters

**Rationale:** Prevent header injection attacks.

### REQ-HEADER-SEC-003: Configurable Headers

**Default header names:**
- `X-Auth-User`
- `X-Auth-Role`
- `X-Auth-Method`

**Allow customization** via config to prevent conflicts with existing headers.

## Test Scenarios

### Scenario 1: Basic Auth Headers
```toml
[[basic_auth]]
name = "admin"
user = "admin"
pass = "secret"
roles = ["admin", "user"]
```

**Request:** `Authorization: Basic YWRtaW46c2VjcmV0`

**Response Headers:**
```
X-Auth-Method: basic
X-Auth-User: admin
X-Auth-Role: admin,user
```

### Scenario 2: JWT Headers with Metadata
```toml
[jwt]
secret = "secret"
issuer = "auth-service"

[headers]
include_jwt_metadata = true
```

**JWT Claims:** `{sub: "user123", iss: "auth-service", role: "admin"}`

**Response Headers:**
```
X-Auth-Method: jwt
X-Auth-User: user123
X-Auth-Role: jwt,admin
X-Auth-Issuer: auth-service
```

### Scenario 3: Authorization Transformation
```toml
[[route_policy]]
name = "api-transform"
host = "api.example.com"
inject_authorization = "Bearer sk-upstream-abc123"
```

**Request:** `Authorization: Basic YWRtaW46c2VjcmV0`

**Response Headers:**
```
Authorization: Bearer sk-upstream-abc123
X-Auth-Method: basic
X-Auth-User: admin
```

**Upstream Receives:** `Authorization: Bearer sk-upstream-abc123`

### Scenario 4: Custom Headers
```toml
[headers]
extra_headers = ["X-Auth-Timestamp", "X-Auth-Route"]
```

**Request:** `GET api.example.com/api/users`

**Response Headers:**
```
X-Auth-Timestamp: 1738560000
X-Auth-Route: api.example.com/api/users
```

### Scenario 5: Anonymous Access
```toml
[[route_policy]]
name = "public"
path_prefix = "/public"
allow_anonymous = true
```

**Request:** `GET /public/status` (no auth)

**Response Headers:**
```
X-Auth-Method: anonymous
```

**Note:** `X-Auth-User` and `X-Auth-Role` not set for anonymous.

### Scenario 6: Custom Header Names
```toml
[headers]
user_header = "X-Forwarded-User"
role_header = "X-User-Roles"
method_header = "X-Auth-Type"
```

**Response Headers:**
```
X-Auth-Type: basic
X-Forwarded-User: admin
X-User-Roles: admin,user
```

### Scenario 7: No Roles
```toml
[[bearer_token]]
name = "token1"
token = "abc123"
roles = []  # No roles
```

**Response Headers:**
```
X-Auth-Method: bearer
X-Auth-Role: 
```

**Note:** Empty `X-Auth-Role` header (not omitted).

## Traefik Configuration

### Minimal Setup
```yaml
- "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
- "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
```

### With Authorization Transformation
```yaml
- "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=Authorization,X-Auth-User,X-Auth-Role"
```

**Important:** Must include `Authorization` in `authResponseHeaders` for transformation to work.

### Trust Forward Headers
```yaml
- "traefik.http.middlewares.auth.forwardauth.trustForwardHeader=false"
```

**Recommendation:** Set to `false` unless you control all proxies in the chain.

### Never Use forwardBody
```yaml
# ❌ DO NOT USE
- "traefik.http.middlewares.auth.forwardauth.forwardBody=true"
```

**Reason:** Breaks streaming (SSE, WebSockets) and adds latency.

## Configuration Validation

### VALID-HEADER-001: Valid Header Names
- Header names must match regex: `^[A-Za-z][A-Za-z0-9-]*$`
- No spaces or special characters except hyphen

### VALID-HEADER-002: No Reserved Headers
- Cannot override: `Host`, `Content-Length`, `Transfer-Encoding`
- Warning if overriding: `Authorization` (only valid via `inject_authorization`)

### VALID-HEADER-003: Unique Extra Headers
- No duplicates in `extra_headers` array

## Logging

**Log format:**
```
[Auth] GET api.example.com/api/users
  Method: basic
  User: admin
  Roles: admin,user
  Headers: X-Auth-User, X-Auth-Role, X-Auth-Method
  Injected: (none)
```

**Do not log:**
- Header values (may contain sensitive info)
- Authorization header content
- Passwords or tokens
