# Changelog

[简体中文](CHANGELOG_ZH.md) | English

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of tiny-auth
- **Security**: Trusted proxies configuration to prevent X-Forwarded-* header spoofing
  - Configure `server.trusted_proxies` with IP/CIDR list
  - Only accepts X-Forwarded-Host/For/Method from trusted sources
  - Logs warning when untrusted proxy detected
  - Supports IPv4, IPv6, single IPs, and CIDR ranges
- Multiple authentication methods support:
  - Basic Auth with constant-time password comparison
  - Bearer Token (static tokens)
  - API Key (via Authorization header or X-Api-Key)
  - JWT validation with HS256
- Route-based policy control:
  - Host matching (exact and wildcard `*.example.com`)
  - Path prefix matching
  - HTTP method matching
  - Anonymous access support
  - Role requirements (all/any)
  - Authentication method whitelist
- Header injection capabilities:
  - Standard headers (User/Role/Method)
  - Custom headers (Timestamp/Route)
  - JWT metadata headers
  - Authorization header transformation
- Configuration management:
  - TOML configuration format
  - Environment variable support (`env:VAR_NAME` syntax)
  - Configuration hot reload (SIGHUP signal)
  - Configuration validation command
  - File permission checks
- CLI commands:
  - `server` - Start authentication service
  - `validate` - Validate configuration file
  - `version` - Show version information
- Docker support:
  - Multi-architecture images (amd64, arm64, arm/v7)
  - Docker Hub and GitHub Container Registry
  - Docker Compose examples
  - Health check configuration
- Development tools:
  - justfile with 18 tasks
  - Makefile support
  - GoReleaser configuration
  - golangci-lint configuration
- Documentation:
  - OpenSpec specification documents
  - Chinese and English README
  - Complete configuration examples
  - Traefik integration guide

### Security
- **CRITICAL FIX**: Fixed jwt_only policy bypass vulnerability (CVE-level)
  - jwt_only = true now correctly rejects non-JWT authentication
  - Added comprehensive test coverage for policy checks
- **NEW**: Trusted proxies validation
  - Prevents X-Forwarded-* header spoofing attacks
  - Configurable via server.trusted_proxies
  - Defaults to accepting all (backward compatible, but warns in logs)
- Constant-time comparison for all credential validation
- Header value sanitization to prevent injection attacks
- Configuration file permission validation
- Weak password warnings
- No secrets in logs
- **IMPROVED**: Structured audit logging with zap
  - JSON format for production (parseable by ELK/Datadog)
  - Request ID for distributed tracing
  - Real client IP validation via trusted_proxies
  - Performance metrics (latency tracking)
  - Security event tracking (failed auth with reasons)

## [0.1.0] - TBD

Initial release.

[Unreleased]: https://github.com/nerdneilsfield/tiny-auth/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/nerdneilsfield/tiny-auth/releases/tag/v0.1.0
