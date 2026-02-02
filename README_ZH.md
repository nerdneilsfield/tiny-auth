# 🔐 tiny-auth

<div align="center">

**史上最轻量级的 Traefik ForwardAuth 认证服务！🚀**

_一个配置文件搞定所有认证需求，妈妈再也不用担心我的 API 安全了！_

[![Go Version](https://img.shields.io/github/go-mod/go-version/nerdneilsfield/tiny-auth?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/nerdneilsfield/tiny-auth?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/nerdneilsfield/tiny-auth?style=flat-square&logo=github)](https://github.com/nerdneilsfield/tiny-auth/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/nerdneils/tiny-auth?style=flat-square&logo=docker)](https://hub.docker.com/r/nerdneils/tiny-auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/nerdneilsfield/tiny-auth?style=flat-square)](https://goreportcard.com/report/github.com/nerdneilsfield/tiny-auth)
[![Build Status](https://img.shields.io/github/actions/workflow/status/nerdneilsfield/tiny-auth/goreleaser.yml?style=flat-square&logo=github-actions)](https://github.com/nerdneilsfield/tiny-auth/actions)

[English](README.md) | 简体中文

</div>

---

## ✨ 为什么选择 tiny-auth？

> 💡 **一句话总结**：如果你在用 Traefik 做反向代理，还在为认证头疼，那 tiny-auth 就是你的救星！

### 🎯 核心亮点

- 🪶 **轻到飞起**：单个二进制，零依赖，5MB 不到
- ⚡ **快如闪电**：Fiber 框架加持，轻松处理 1000+ req/s
- 🔒 **安全至上**：常量时间比较防时序攻击，配置文件权限检查
- 🎨 **灵活配置**：TOML 格式，支持环境变量，热重载不重启
- 🌈 **多种认证**：Basic Auth / Bearer Token / API Key / JWT 一网打尽
- 🎯 **精准控制**：基于 host/path/method 的路由策略，想怎么玩就怎么玩
- 🔄 **Header 注入**：客户端用 Basic Auth，上游收到 Bearer Token？没问题！
- 📊 **开箱即用**：健康检查、调试端点、配置验证，该有的都有

### 🆚 对比其他方案

| 特性 | tiny-auth | Traefik BasicAuth | OAuth2 Proxy | Authelia |
|------|-----------|-------------------|--------------|----------|
| 二进制大小 | ~5MB | N/A | ~20MB | ~30MB |
| 多种认证方式 | ✅ | ❌ | ✅ | ✅ |
| 路由级策略 | ✅ | ❌ | ❌ | ✅ |
| Header 转换 | ✅ | ❌ | ⚠️ | ❌ |
| 配置热重载 | ✅ | ❌ | ❌ | ✅ |
| 环境变量支持 | ✅ | ✅ | ✅ | ✅ |
| 复杂度 | ⭐ | ⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |

---

## 📦 Docker 镜像

tiny-auth 提供两个镜像源，国内外均可快速拉取：

| 镜像源 | 地址 | 推荐地区 |
|--------|------|----------|
| 🐳 **Docker Hub** | `nerdneils/tiny-auth:latest` | 🌍 国际 |
| 📦 **GitHub CR** | `ghcr.io/nerdneilsfield/tiny-auth:latest` | 🌏 国内（可能较慢） |

**支持的架构**：
- `linux/amd64` - x86_64
- `linux/arm64` - ARM64 / Apple Silicon
- `linux/arm/v7` - ARMv7

## 🚀 快速开始

### 方式一：Docker（推荐）

```bash
# 1. 创建配置文件
cat > config.toml << 'EOF'
[server]
port = "8080"

[[basic_auth]]
name = "admin"
user = "admin"
pass = "supersecret"
roles = ["admin"]
EOF

# 2. 运行（两个镜像源任选其一）
# Docker Hub
docker run -d \
  --name tiny-auth \
  -p 8080:8080 \
  -v $(pwd)/config.toml:/root/config.toml:ro \
  nerdneils/tiny-auth:latest

# 或者使用 GitHub Container Registry
docker run -d \
  --name tiny-auth \
  -p 8080:8080 \
  -v $(pwd)/config.toml:/root/config.toml:ro \
  ghcr.io/nerdneilsfield/tiny-auth:latest

# 3. 测试
curl -u admin:supersecret http://localhost:8080/auth
# → 200 OK ✅
```

### 方式二：二进制下载

```bash
# 下载最新版本
wget https://github.com/nerdneilsfield/tiny-auth/releases/latest/download/tiny-auth_linux_amd64.tar.gz
tar -xzf tiny-auth_linux_amd64.tar.gz

# 运行
./tiny-auth server --config config.toml
```

### 方式三：从源码编译

```bash
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth
just build  # 或者 make build
./tiny-auth server
```

---

## 🎨 配置示例

<details>
<summary><b>📖 完整配置示例（点击展开）</b></summary>

```toml
# ===== 服务器配置 =====
[server]
port = "8080"
auth_path = "/auth"
health_path = "/health"

# ===== 日志配置 =====
[logging]
format = "json"  # 或 "text"
level = "info"   # debug/info/warn/error

# ===== Basic Auth =====
[[basic_auth]]
name = "admin-user"
user = "admin"
pass = "supersecret"        # 支持 env:PASSWORD 从环境变量读取
roles = ["admin", "user"]

# ===== Bearer Token =====
[[bearer_token]]
name = "api-token"
token = "env:API_TOKEN"     # 从环境变量读取
roles = ["api", "service"]

# ===== API Key =====
[[api_key]]
name = "prod-key"
key = "ak_prod_xxx"
roles = ["admin"]

# ===== JWT 验证 =====
[jwt]
secret = "your-256-bit-secret-key-must-be-32-chars"
issuer = "auth-service"
audience = "api"

# ===== 路由策略 =====
# 公共 API 允许匿名
[[route_policy]]
name = "public"
path_prefix = "/public"
allow_anonymous = true

# 管理面板只允许 admin
[[route_policy]]
name = "admin"
host = "admin.example.com"
allowed_basic_names = ["admin-user"]
require_all_roles = ["admin"]

# Webhook 端点并注入上游 token
[[route_policy]]
name = "webhook"
host = "hooks.example.com"
path_prefix = "/webhook"
allowed_bearer_names = ["api-token"]
inject_authorization = "Bearer upstream-token-123"
```

</details>

<details>
<summary><b>🔑 环境变量语法</b></summary>

在配置文件中使用 `env:VAR_NAME` 从环境变量读取敏感信息：

```toml
[[basic_auth]]
pass = "env:ADMIN_PASSWORD"

[jwt]
secret = "env:JWT_SECRET"
```

启动时设置环境变量：

```bash
export ADMIN_PASSWORD="my-secure-password"
export JWT_SECRET="my-jwt-secret-key-32-chars-long"
./tiny-auth server
```

</details>

---

## 🔌 Traefik 集成

### Docker Compose 完整示例

```yaml
version: '3.8'

services:
  # tiny-auth 认证服务
  tiny-auth:
    image: nerdneils/tiny-auth:latest  # 或使用 ghcr.io/nerdneilsfield/tiny-auth:latest
    volumes:
      - ./config.toml:/root/config.toml:ro
    networks:
      - traefik

  # Traefik 反向代理
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

  # 受保护的服务（示例）
  whoami:
    image: traefik/whoami
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.whoami.rule=Host(`whoami.localhost`)"
      
      # 配置 ForwardAuth 中间件
      - "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
      - "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
      
      # 应用中间件
      - "traefik.http.routers.whoami.middlewares=auth@docker"
    networks:
      - traefik

networks:
  traefik:
```

### 关键配置说明

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `address` | tiny-auth 的 /auth 端点地址 | `http://tiny-auth:8080/auth` |
| `authResponseHeaders` | 要注入到上游的 headers | `X-Auth-User,X-Auth-Role` |
| `trustForwardHeader` | 是否信任 X-Forwarded-* | `false`（推荐） |

⚠️ **重要**：千万别启用 `forwardBody=true`，会破坏 SSE/WebSocket！

---

## 🎯 使用场景

### 场景 1：多环境 API 认证

```toml
# 开发环境用 Basic Auth
[[basic_auth]]
name = "dev"
user = "dev"
pass = "dev123"
roles = ["developer"]

# 生产环境用 Bearer Token
[[bearer_token]]
name = "prod"
token = "env:PROD_TOKEN"
roles = ["admin", "service"]

# 路由策略
[[route_policy]]
name = "dev-api"
host = "dev-api.example.com"
allowed_basic_names = ["dev"]

[[route_policy]]
name = "prod-api"
host = "api.example.com"
allowed_bearer_names = ["prod"]
```

### 场景 2：认证转换（客户端 Basic → 上游 Bearer）

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

客户端用 Basic Auth 访问，上游服务收到的是 Bearer Token！

### 场景 3：微服务内部认证

```toml
# 服务间通信用 API Key
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

## 🐳 Docker 使用指南

<details>
<summary><b>🏃 快速启动完整环境</b></summary>

使用我们提供的完整示例（包含 Traefik + tiny-auth + 5个示例服务）：

```bash
# 1. 克隆仓库
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth/examples

# 2. 准备配置
cp config-full.toml config.toml
cp .env.example .env
# 编辑 .env 文件，设置你的密码

# 3. 启动服务
docker-compose -f docker-compose-full.yml up -d

# 4. 测试
curl -u admin:your-password http://whoami-basic.localhost/
curl http://public.localhost/public/
curl -H "Authorization: Bearer your-token" http://api.localhost/

# 5. 查看日志
docker-compose -f docker-compose-full.yml logs -f tiny-auth

# 6. 停止服务
docker-compose -f docker-compose-full.yml down
```

</details>

<details>
<summary><b>🔧 自定义 Docker Compose</b></summary>

在你的项目中集成 tiny-auth：

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
<summary><b>🔄 配置热重载</b></summary>

修改配置文件后，无需重启容器：

```bash
# 方式一：发送 SIGHUP 信号
docker kill --signal=SIGHUP tiny-auth

# 方式二：使用 docker-compose
docker-compose kill -s SIGHUP tiny-auth

# 查看重载日志
docker logs tiny-auth --tail 20
# → "♻️  Configuration reloaded"
```

</details>

<details>
<summary><b>🏗️ 从源码构建镜像</b></summary>

```bash
# 克隆仓库
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth

# 构建镜像
docker build -t my-tiny-auth:latest .

# 或使用 justfile
just docker-build

# 运行
docker run -d -p 8080:8080 \
  -v $(pwd)/config.toml:/root/config.toml:ro \
  my-tiny-auth:latest
```

</details>

---

## 📊 运维管理

### 配置热重载

```bash
# 发送 SIGHUP 信号重新加载配置
kill -HUP $(pidof tiny-auth)

# 或者使用 Docker
docker kill --signal=SIGHUP tiny-auth
```

无需重启，配置即时生效！✨

### 配置验证

```bash
# 验证配置文件
tiny-auth validate config.toml

# 输出示例
✅ Configuration is valid

📋 Configuration Summary:
✓ Server: port 8080
✓ Basic Auth: 2 users configured
✓ Bearer Tokens: 1 tokens configured
✓ Route Policies: 3 policies configured

⚠ Warning: config.toml has insecure permissions 644
⚠ Recommendation: chmod 0600 config.toml
```

### 健康检查

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

### 调试端点（可选）

先在配置中启用：

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

⚠️ **不要在公网暴露该端点**，仅在可信网络中使用。

---

## 🔒 安全最佳实践

### ✅ 必须做的

1. **⚠️ 配置可信代理（非常重要！）**
   
   **为什么**：防止攻击者伪造 `X-Forwarded-*` headers 绕过策略。

   **重要**：反向代理/负载均衡必须清理或覆盖客户端伪造的 `X-Forwarded-*` 头部。  
   否则即使配置了 `trusted_proxies` 也可能被绕过。
   
   ```toml
   [server]
   # ✅ 生产环境：只信任你的反向代理
   trusted_proxies = ["172.16.0.0/12"]  # Docker 网络
   
   # ❌ 不安全：空列表接受任何来源的 headers
   # trusted_proxies = []
   ```
   
   **示例配置**：
   - Docker Compose: `["172.16.0.0/12"]`
   - Kubernetes: `["10.0.0.0/8"]`
   - 特定 IP: `["192.168.1.100"]`
   - 多个网段: `["172.16.0.0/12", "192.168.1.0/24"]`
   
   **不配置会怎样**：
   ```bash
   # 攻击者可以伪造 host 绕过策略
   curl -H "X-Forwarded-Host: admin.internal.com" \
        http://your-tiny-auth:8080/auth
   # 没有 trusted_proxies: ✅ 允许通过（策略被绕过！）
   # 配置了 trusted_proxies:  ❌ 拒绝访问（headers 被忽略）
   ```

2. **配置文件权限**
   ```bash
   chmod 0600 config.toml  # 只有所有者可读写
   ```

3. **使用环境变量存储敏感信息**
   ```toml
   pass = "env:ADMIN_PASSWORD"  # ✅
   pass = "plaintext123"        # ❌
   ```

4. **强密码策略**
   - 至少 12 个字符
   - 包含大小写字母、数字、特殊字符

5. **JWT 密钥长度**
   - 至少 32 字符（256 bits）

6. **生产环境启用 JSON 日志**
   ```toml
   [logging]
   format = "json"  # 结构化，可搜索
   level = "info"
   ```
   
   **结构化日志包含**：
   - `request_id`：跨服务追踪
   - `client_ip`：真实客户端 IP（通过 trusted_proxies 验证）
   - `auth_method`：哪种认证方式成功
   - `latency`：性能监控
   - `reason`：认证失败原因

### ⚠️ 注意事项

- 不要在公网暴露调试端点 `/debug/config`
- 定期轮换 tokens 和 API keys
- 使用 HTTPS（Traefik 配置 TLS）
- 审查认证日志，监控异常访问

---

## 🛠️ 开发指南

<details>
<summary><b>🔨 本地开发</b></summary>

### 前置要求

- Go 1.23+
- [just](https://github.com/casey/just) 或 make

### 常用命令

```bash
# 克隆仓库
git clone https://github.com/nerdneilsfield/tiny-auth.git
cd tiny-auth

# 安装依赖
just deps

# 编译
just build

# 运行测试
just test

# 代码检查
just lint

# 格式化代码
just fmt

# 完整检查（测试 + lint）
just check

# 查看所有命令
just --list
```

### 目录结构

```
tiny-auth/
├── cmd/                # CLI 命令
│   ├── root.go        # 根命令
│   ├── server.go      # 服务器命令
│   ├── validate.go    # 配置验证命令
│   └── version.go     # 版本信息命令
├── internal/          # 内部包
│   ├── config/        # 配置管理
│   ├── auth/          # 认证逻辑
│   ├── policy/        # 策略匹配
│   └── server/        # HTTP 服务器
├── openspec/          # OpenSpec 规范文档
└── main.go            # 入口文件
```

</details>

<details>
<summary><b>🧪 测试</b></summary>

```bash
# 运行所有测试
just test

# 生成覆盖率报告
just test-coverage
open coverage.html

# 竞态检测
go test -race ./...
```

当前测试覆盖率目标：>80%

</details>

---

## 📚 文档

- [完整配置参考](openspec/changes/initial-implementation/specs/04-configuration.md)
- [认证方法详解](openspec/changes/initial-implementation/specs/01-authentication.md)
- [路由策略详解](openspec/changes/initial-implementation/specs/02-route-policies.md)
- [Header 注入详解](openspec/changes/initial-implementation/specs/03-header-injection.md)
- [技术设计文档](openspec/changes/initial-implementation/design.md)

---

## 🤝 贡献指南

欢迎所有形式的贡献！

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

提交前请运行：

```bash
just pre-commit  # 格式化 + 检查
```

---

## 📝 变更日志

查看 [CHANGELOG.md](CHANGELOG.md) 了解版本历史。

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

---

## 🙏 致谢

- [Fiber](https://github.com/gofiber/fiber) - 高性能 Web 框架
- [Traefik](https://github.com/traefik/traefik) - 现代化反向代理
- [golang-jwt](https://github.com/golang-jwt/jwt) - JWT 实现
- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML 解析器

---

## 💬 社区与支持

- 🐛 [问题反馈](https://github.com/nerdneilsfield/tiny-auth/issues)
- 💡 [功能建议](https://github.com/nerdneilsfield/tiny-auth/discussions)
- 📧 联系作者：dengqi935@gmail.com

---

<div align="center">

**⭐ 如果觉得有用，请给个 Star 支持一下！⭐**

Made with ❤️ by [dengqi](https://github.com/nerdneilsfield)

</div>
