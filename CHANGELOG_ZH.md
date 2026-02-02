# Changelog

[English](CHANGELOG.md) | 简体中文

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of tiny-auth
- **安全功能**: 可信代理配置，防止 X-Forwarded-* header 伪造
  - 通过 `server.trusted_proxies` 配置 IP/CIDR 列表
  - 仅接受来自可信来源的 X-Forwarded-Host/For/Method
  - 检测到不可信代理时记录警告
  - 支持 IPv4、IPv6、单个 IP 和 CIDR 范围
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
- **重大修复**: 修复 jwt_only 策略绕过漏洞（CVE 级别）
  - jwt_only = true 现在正确拒绝非 JWT 认证
  - 添加了策略检查的完整测试覆盖
- **新功能**: 可信代理验证
  - 防止 X-Forwarded-* header 伪造攻击
  - 通过 server.trusted_proxies 配置
  - 默认接受所有（向后兼容，但会在日志中警告）
- 所有凭证验证使用常量时间比较
- Header 值清理，防止注入攻击
- 配置文件权限验证
- 弱密码警告
- 日志中不包含敏感信息
- **改进**: 使用 zap 的结构化审计日志
  - 生产环境 JSON 格式（可被 ELK/Datadog 解析）
  - Request ID 用于分布式追踪
  - 通过 trusted_proxies 验证的真实客户端 IP
  - 性能指标（延迟追踪）
  - 安全事件追踪（认证失败及原因）

## [0.1.0] - TBD

Initial release.

[Unreleased]: https://github.com/nerdneilsfield/tiny-auth/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/nerdneilsfield/tiny-auth/releases/tag/v0.1.0
