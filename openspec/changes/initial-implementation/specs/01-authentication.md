# Spec: Authentication Methods

## Overview

Define how tiny-auth validates credentials for each supported authentication method.

## Requirements

### REQ-AUTH-001: Basic Authentication

**Given** a request with `Authorization: Basic <base64>` header  
**When** credentials match a configured `[[basic_auth]]` entry  
**Then** authentication succeeds and returns user's roles

**Details:**
- Decode base64 to get `username:password`
- Look up username in auth store
- Compare password using `crypto/subtle.ConstantTimeCompare`
- Return username and associated roles

**Example:**
```toml
[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "supersecret"
roles = ["admin", "user"]
```

```bash
curl -u admin:supersecret https://api.example.com/endpoint
# → Authorization: Basic YWRtaW46c3VwZXJzZWNyZXQ=
```

### REQ-AUTH-002: Bearer Token

**Given** a request with `Authorization: Bearer <token>` header  
**When** token matches a configured `[[bearer_token]]` entry  
**Then** authentication succeeds and returns token's roles

**Details:**
- Extract token from `Bearer ` prefix
- Look up token in auth store using constant-time comparison
- Distinguish between JWT (3 segments with `.`) and static token
- Return token name and associated roles

**Example:**
```toml
[[bearer_token]]
name = "api-token"
token = "sk-prod-abc123xyz789"
roles = ["api", "service"]
```

```bash
curl -H "Authorization: Bearer sk-prod-abc123xyz789" https://api.example.com/endpoint
```

### REQ-AUTH-003: JWT Validation

**Given** a request with `Authorization: Bearer <jwt>` header  
**When** JWT has valid signature and claims  
**Then** authentication succeeds and returns JWT subject as user

**Details:**
- Parse JWT using golang-jwt/jwt/v5
- Validate HS256 signature using configured secret
- Verify `iss` (issuer) matches config if specified
- Verify `aud` (audience) matches config if specified
- Extract `sub` (subject) as username
- Extract `role` claim (if present) and add to roles array
- Default role: `["jwt"]`

**Example:**
```toml
[jwt]
secret = "your-256-bit-secret-key-here"
issuer = "auth-service"
audience = "api"
```

```bash
# JWT payload: {"sub": "user123", "role": "admin", "iss": "auth-service", "aud": "api"}
curl -H "Authorization: Bearer eyJhbGc..." https://api.example.com/endpoint
```

**Edge Cases:**
- JWT with 3 segments but invalid signature → reject
- Bearer token that looks like JWT → try JWT first, fall back to static token
- JWT without `sub` claim → reject
- Expired JWT → reject
- JWT with wrong issuer/audience → reject

### REQ-AUTH-004: API Key (Authorization Header)

**Given** a request with `Authorization: ApiKey <key>` header  
**When** key matches a configured `[[api_key]]` entry  
**Then** authentication succeeds and returns key's roles

**Example:**
```toml
[[api_key]]
name = "prod-key"
key = "ak_prod_xxx_secret"
roles = ["admin"]
```

```bash
curl -H "Authorization: ApiKey ak_prod_xxx_secret" https://api.example.com/endpoint
```

### REQ-AUTH-005: API Key (X-Api-Key Header)

**Given** a request with `X-Api-Key: <key>` header  
**When** key matches a configured `[[api_key]]` entry  
**Then** authentication succeeds and returns key's roles

**Example:**
```bash
curl -H "X-Api-Key: ak_prod_xxx_secret" https://api.example.com/endpoint
```

## Authentication Priority

When multiple methods are possible, try in this order:

1. **JWT** (if secret configured and Bearer header present with 3 segments)
2. **Bearer Token** (if Bearer header present and not JWT)
3. **Basic Auth** (if Basic header present)
4. **API Key Authorization** (if `Authorization: ApiKey` present)
5. **API Key Header** (if `X-Api-Key` present)

**Rationale:** JWT is most specific, Basic is standard, API Key is fallback.

## Authentication Result

Successful authentication returns:
```go
type AuthResult struct {
    Method   string              // "basic", "bearer", "apikey", "jwt"
    Name     string              // Config name (e.g., "admin-user")
    User     string              // Username or subject
    Roles    []string            // Associated roles
    Metadata map[string]string   // Extra info (e.g., JWT issuer)
}
```

Failed authentication returns `nil`.

## Security Requirements

### REQ-AUTH-SEC-001: Constant-Time Comparison

**Must use** `crypto/subtle.ConstantTimeCompare` for:
- Password comparison (Basic Auth)
- Token comparison (Bearer, API Key)

**Rationale:** Prevent timing attacks that could leak credential information.

### REQ-AUTH-SEC-002: No Secrets in Logs

**Must not** log:
- Passwords
- Tokens
- API keys
- JWT signatures
- Authorization header values

**May log:**
- Usernames
- Token/key names
- Authentication method used
- Success/failure status

### REQ-AUTH-SEC-003: Secure Defaults

**Default roles:**
- Basic Auth: `["user"]`
- Bearer Token: `["service"]`
- API Key: `["api"]`
- JWT: `["jwt"]` + extracted role claim

**Default behavior:**
- If no route policy matches → require authentication (deny anonymous)
- If policy exists but no roles required → allow any authenticated user
- If multiple auth methods provided → use first valid one

## Test Scenarios

### Scenario 1: Valid Basic Auth
```
Request: Authorization: Basic YWRtaW46c2VjcmV0
Config: user=admin, pass=secret, roles=["admin"]
Expected: Success, user="admin", roles=["admin"]
```

### Scenario 2: Invalid Password
```
Request: Authorization: Basic YWRtaW46d3Jvbmc=
Config: user=admin, pass=secret
Expected: Failure (401)
```

### Scenario 3: Valid JWT
```
Request: Authorization: Bearer eyJ...validJWT
Config: jwt.secret="key", jwt.issuer="auth"
JWT Claims: {sub: "user123", iss: "auth", role: "admin"}
Expected: Success, user="user123", roles=["jwt", "admin"]
```

### Scenario 4: Invalid JWT Signature
```
Request: Authorization: Bearer eyJ...invalidSignature
Expected: Failure (401)
```

### Scenario 5: Static Bearer Token
```
Request: Authorization: Bearer sk-prod-abc123
Config: token="sk-prod-abc123", roles=["api"]
Expected: Success, roles=["api"]
```

### Scenario 6: API Key in Authorization Header
```
Request: Authorization: ApiKey ak_prod_xxx
Config: key="ak_prod_xxx", roles=["admin"]
Expected: Success, roles=["admin"]
```

### Scenario 7: API Key in Custom Header
```
Request: X-Api-Key: ak_prod_xxx
Config: key="ak_prod_xxx", roles=["admin"]
Expected: Success, roles=["admin"]
```

### Scenario 8: Multiple Auth Methods (Priority)
```
Request: 
  Authorization: Bearer eyJ...validJWT
  X-Api-Key: ak_prod_xxx
Config: Both JWT and API Key configured
Expected: JWT takes priority, API Key ignored
```

### Scenario 9: No Authentication Header
```
Request: (no Authorization header)
Expected: Failure (401) with WWW-Authenticate challenge
```

### Scenario 10: Unknown Token
```
Request: Authorization: Bearer unknown-token
Config: (token not in config)
Expected: Failure (401)
```

## Configuration Validation

### VALID-AUTH-001: No Empty Credentials
- Basic Auth: `user` and `pass` must be non-empty
- Bearer Token: `token` must be non-empty
- API Key: `key` must be non-empty
- JWT: `secret` must be at least 32 characters (256 bits)

### VALID-AUTH-002: Unique Names
- All `name` fields must be unique within their auth type
- Prevents ambiguity in route policy references

### VALID-AUTH-003: No Duplicate Credentials
- No two Basic Auth configs can have the same `user`
- No two Bearer Token configs can have the same `token`
- No two API Key configs can have the same `key`
- (JWT is singleton, no duplicates possible)

## Error Messages

```
401 Unauthorized:
{
  "error": "Unauthorized",
  "timestamp": 1738560000
}

WWW-Authenticate: Basic realm="api"
WWW-Authenticate: Bearer realm="api"
```

Multiple `WWW-Authenticate` headers inform client which methods are accepted.
