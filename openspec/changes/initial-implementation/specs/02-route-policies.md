# Spec: Route Policies

## Overview

Define how tiny-auth matches and enforces route-based authentication policies.

## Requirements

### REQ-POLICY-001: Route Policy Matching

**Given** an incoming request with host, path, and method  
**When** evaluating route policies  
**Then** return the first matching policy or `nil` if no match

**Matching Rules:**

1. **Host Matching**
   - Exact match: `host = "api.example.com"`
   - Wildcard suffix: `host = "*.example.com"` matches `api.example.com`, `admin.example.com`
   - Empty host: matches all hosts

2. **Path Matching**
   - Prefix match: `path_prefix = "/api"` matches `/api/users`, `/api/v1/posts`
   - Empty path_prefix: matches all paths
   - Case-sensitive

3. **Method Matching**
   - Exact match: `method = "GET"` matches only GET requests
   - Case-insensitive
   - Empty method: matches all methods

**Priority:** First policy that matches all specified fields wins.

**Example:**
```toml
[[route_policy]]
name = "admin-api"
host = "admin.example.com"
path_prefix = "/api"
method = "POST"
# Matches: POST admin.example.com/api/users
# Doesn't match: GET admin.example.com/api/users (different method)
# Doesn't match: POST api.example.com/api/users (different host)
```

### REQ-POLICY-002: Anonymous Access

**Given** a route policy with `allow_anonymous = true`  
**When** request matches that policy  
**Then** allow access without authentication

**Example:**
```toml
[[route_policy]]
name = "public-api"
host = "api.example.com"
path_prefix = "/public"
allow_anonymous = true
```

```bash
# No authentication required
curl https://api.example.com/public/status
# → 200 OK
```

### REQ-POLICY-003: Restrict Authentication Methods

**Given** a route policy with allowed method names  
**When** authentication succeeds  
**Then** verify auth method name is in allowed list

**Fields:**
- `allowed_basic_names`: Only allow specific Basic Auth configs
- `allowed_bearer_names`: Only allow specific Bearer Token configs
- `allowed_api_key_names`: Only allow specific API Key configs
- `jwt_only`: Only allow JWT authentication

**Example:**
```toml
[[route_policy]]
name = "webhook-endpoint"
host = "hooks.example.com"
path_prefix = "/webhook"
allowed_bearer_names = ["webhook-token"]
# Only accepts the bearer token named "webhook-token"
```

**Behavior:**
- If `allowed_*` lists are empty → allow all of that type
- If `allowed_*` lists are specified → only allow listed names
- If `jwt_only = true` → only JWT accepted (other lists ignored)

### REQ-POLICY-004: Role Requirements

**Given** a route policy with role requirements  
**When** authentication succeeds with roles  
**Then** verify user has required roles

**Fields:**
- `require_all_roles`: User must have ALL listed roles
- `require_any_role`: User must have AT LEAST ONE listed role

**Example:**
```toml
[[route_policy]]
name = "admin-panel"
host = "admin.example.com"
require_all_roles = ["admin"]
# User must have "admin" role

[[route_policy]]
name = "mixed-access"
host = "mixed.example.com"
require_any_role = ["admin", "service"]
# User must have either "admin" OR "service" role
```

**Edge Cases:**
- If both `require_all_roles` and `require_any_role` specified → must satisfy BOTH conditions
- If neither specified → any authenticated user allowed
- If user has no roles → fail requirements check

### REQ-POLICY-005: No Policy Match

**Given** no route policy matches the request  
**When** evaluating authentication  
**Then** require authentication but accept any valid method

**Behavior:**
- Try all configured authentication methods
- Any valid credential accepted
- No specific role requirements
- Equivalent to: global default auth policy

## Policy Checking Algorithm

```go
func checkPolicy(policy *RoutePolicy, result *AuthResult) bool {
    if policy == nil {
        return true  // No policy = accept any valid auth
    }
    
    // Check method whitelist
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
        if policy.JWTOnly {
            // JWT is allowed
        }
    }
    
    // Check role requirements
    if len(policy.RequireAllRoles) > 0 {
        for _, required := range policy.RequireAllRoles {
            if !contains(result.Roles, required) {
                return false
            }
        }
    }
    
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

## Test Scenarios

### Scenario 1: Exact Host Match
```toml
[[route_policy]]
name = "exact"
host = "api.example.com"
```
- ✅ `api.example.com` → matches
- ❌ `admin.example.com` → no match
- ❌ `api.example.com.evil.com` → no match

### Scenario 2: Wildcard Host Match
```toml
[[route_policy]]
name = "wildcard"
host = "*.example.com"
```
- ✅ `api.example.com` → matches
- ✅ `admin.example.com` → matches
- ✅ `foo.bar.example.com` → matches (suffix match)
- ❌ `example.com` → no match (no subdomain)

### Scenario 3: Path Prefix Match
```toml
[[route_policy]]
name = "api-routes"
path_prefix = "/api"
```
- ✅ `/api/users` → matches
- ✅ `/api/v1/posts` → matches
- ✅ `/api` → matches (exact)
- ❌ `/public/api` → no match (not prefix)

### Scenario 4: Method Match
```toml
[[route_policy]]
name = "post-only"
method = "POST"
```
- ✅ `POST /anything` → matches
- ❌ `GET /anything` → no match
- ❌ `PUT /anything` → no match

### Scenario 5: Combined Match
```toml
[[route_policy]]
name = "specific"
host = "admin.example.com"
path_prefix = "/api/admin"
method = "POST"
```
- ✅ `POST admin.example.com/api/admin/users` → matches
- ❌ `POST admin.example.com/api/users` → no match (wrong path)
- ❌ `GET admin.example.com/api/admin/users` → no match (wrong method)
- ❌ `POST api.example.com/api/admin/users` → no match (wrong host)

### Scenario 6: Anonymous Route
```toml
[[route_policy]]
name = "public"
path_prefix = "/public"
allow_anonymous = true
```
- Request: `GET /public/status` (no auth header)
- Expected: 200 OK (no authentication required)

### Scenario 7: Restricted Auth Method
```toml
[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "secret"
roles = ["admin"]

[[basic_auth]]
name = "dev-user"
user = "dev"
pass = "secret"
roles = ["developer"]

[[route_policy]]
name = "admin-only"
host = "admin.example.com"
allowed_basic_names = ["admin-user"]
```
- Request: `admin.example.com` with `admin:secret` → ✅ (admin-user)
- Request: `admin.example.com` with `dev:secret` → ❌ (dev-user not allowed)

### Scenario 8: Role Requirement (All)
```toml
[[basic_auth]]
name = "user1"
user = "user1"
pass = "pass"
roles = ["admin", "dev"]

[[basic_auth]]
name = "user2"
user = "user2"
pass = "pass"
roles = ["admin"]

[[route_policy]]
name = "multi-role"
require_all_roles = ["admin", "dev"]
```
- Auth as `user1` (roles: admin, dev) → ✅
- Auth as `user2` (roles: admin) → ❌ (missing "dev")

### Scenario 9: Role Requirement (Any)
```toml
[[route_policy]]
name = "flexible"
require_any_role = ["admin", "service"]
```
- Auth with roles `["admin"]` → ✅
- Auth with roles `["service"]` → ✅
- Auth with roles `["admin", "service"]` → ✅
- Auth with roles `["user"]` → ❌

### Scenario 10: JWT Only
```toml
[jwt]
secret = "secret"

[[bearer_token]]
name = "static"
token = "token123"
roles = ["api"]

[[route_policy]]
name = "jwt-required"
host = "secure.example.com"
jwt_only = true
```
- Request with valid JWT → ✅
- Request with static Bearer token → ❌ (JWT only)
- Request with Basic Auth → ❌ (JWT only)

### Scenario 11: First Match Wins
```toml
[[route_policy]]
name = "specific"
host = "api.example.com"
path_prefix = "/admin"
require_all_roles = ["admin"]

[[route_policy]]
name = "general"
host = "api.example.com"
allow_anonymous = true
```
- Request: `GET api.example.com/admin/users`
- Matches: "specific" policy (first match)
- Requires: admin role
- Does NOT match: "general" policy (not evaluated)

## Configuration Validation

### VALID-POLICY-001: Policy Name Unique
- All `name` fields in `[[route_policy]]` must be unique

### VALID-POLICY-002: Valid References
- `allowed_basic_names` must reference existing `[[basic_auth]]` names
- `allowed_bearer_names` must reference existing `[[bearer_token]]` names
- `allowed_api_key_names` must reference existing `[[api_key]]` names

### VALID-POLICY-003: Conflicting Settings
- If `allow_anonymous = true`, cannot have role requirements or method restrictions
  (Warning: anonymous access ignores other restrictions)

### VALID-POLICY-004: JWT Only Constraint
- If `jwt_only = true`, cannot have `allowed_basic_names`, `allowed_bearer_names`, or `allowed_api_key_names`
  (Warning: JWT only mode ignores other method restrictions)

## Traefik Integration

Traefik forwards these headers to tiny-auth:
- `X-Forwarded-Host`: Original host (e.g., `api.example.com`)
- `X-Forwarded-Uri`: Original path (e.g., `/api/users`)
- `X-Forwarded-Method`: HTTP method (e.g., `GET`)
- `X-Forwarded-Proto`: Protocol (e.g., `https`)
- `X-Forwarded-For`: Client IP

tiny-auth uses `X-Forwarded-Host`, `X-Forwarded-Uri`, and `X-Forwarded-Method` for policy matching.
