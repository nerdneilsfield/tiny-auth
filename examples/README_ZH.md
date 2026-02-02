# 示例配置

[English](README.md) | 简体中文

本目录包含各种使用场景的配置示例。

## 文件说明

- **docker-compose-full.yml** - 完整的 Traefik + tiny-auth 示例
  - 包含 5 种不同的认证场景
  - 演示多种认证方法
  - 可直接使用的完整环境

- **config-full.toml** - 对应的完整配置文件
  - 包含所有认证方法的配置
  - 演示路由策略用法
  - 使用环境变量存储敏感信息

## 快速开始

### 1. 准备配置文件

```bash
cd examples
cp config-full.toml config.toml
```

### 2. 设置环境变量（可选）

创建 `.env` 文件：

```bash
cat > .env << 'EOF'
ADMIN_PASSWORD=your-secure-admin-password
DEV_PASSWORD=your-dev-password
API_TOKEN=your-api-token-here
JWT_SECRET=your-jwt-secret-key-must-be-at-least-32-chars
EOF
```

### 3. 启动服务

```bash
docker-compose -f docker-compose-full.yml up -d
```

### 4. 测试各种场景

#### 场景 1：Basic Auth

```bash
# 成功 - admin 用户
curl -u admin:your-secure-admin-password http://whoami-basic.localhost/

# 成功 - dev 用户
curl -u dev:your-dev-password http://whoami-basic.localhost/

# 失败 - 错误密码
curl -u admin:wrong http://whoami-basic.localhost/
# → 401 Unauthorized
```

#### 场景 2：Bearer Token

```bash
# 成功
curl -H "Authorization: Bearer your-api-token-here" http://api.localhost/

# 失败 - 无效 token
curl -H "Authorization: Bearer invalid" http://api.localhost/
# → 401 Unauthorized
```

#### 场景 3：公共访问（无需认证）

```bash
# 成功 - 匿名访问
curl http://public.localhost/public/
```

#### 场景 4：管理后台（仅 admin）

```bash
# 成功 - admin 用户
curl -u admin:your-secure-admin-password http://admin.localhost/

# 失败 - dev 用户（没有 admin 角色）
curl -u dev:your-dev-password http://admin.localhost/
# → 401 Unauthorized (policy requirements not met)
```

#### 场景 5：API Key 认证

```bash
# 成功 - 使用 X-Api-Key header
curl -H "X-Api-Key: ak_internal_secret_key" http://internal.localhost/

# 成功 - 使用 Authorization header
curl -H "Authorization: ApiKey ak_internal_secret_key" http://internal.localhost/
```

### 5. 查看认证 Headers

所有成功的请求都会收到注入的认证 headers：

```bash
curl -v -u admin:your-secure-admin-password http://whoami-basic.localhost/

# 响应头包含：
# X-Auth-User: admin
# X-Auth-Role: admin,user
# X-Auth-Method: basic
# X-Auth-Timestamp: 1738560000
```

### 6. 健康检查

```bash
# tiny-auth 健康检查
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

访问 http://traefik.localhost:8080 查看 Traefik Dashboard。

## 配置说明

### 认证流程

```
客户端 → Traefik (检测到需要认证) → tiny-auth (/auth)
                                         ↓
                                    验证凭证
                                         ↓
                                    检查策略
                                         ↓
                          返回 200 + Headers / 401
                                         ↓
          Traefik (注入 Headers) → 上游服务
```

### ForwardAuth 配置要点

```yaml
labels:
  # ForwardAuth 地址
  - "traefik.http.middlewares.auth.forwardauth.address=http://tiny-auth:8080/auth"
  
  # 要注入到上游的 headers
  - "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
  
  # 不信任已有的 X-Forwarded-* headers（安全）
  - "traefik.http.middlewares.auth.forwardauth.trustForwardHeader=false"
  
  # 应用中间件到路由
  - "traefik.http.routers.myservice.middlewares=auth@docker"
```

⚠️ **重要**：
- 不要使用 `forwardBody=true`，会破坏 SSE/WebSocket
- 确保 `authResponseHeaders` 包含你需要的所有 headers
- `trustForwardHeader` 建议设置为 `false`

## 停止服务

```bash
docker-compose -f docker-compose-full.yml down
```

## 故障排查

### 问题：始终返回 401

1. 检查 tiny-auth 日志
   ```bash
   docker logs tiny-auth
   ```

2. 检查配置文件
   ```bash
   docker exec tiny-auth cat /root/config.toml
   ```

3. 验证配置
   ```bash
   docker exec tiny-auth ./tiny-auth validate /root/config.toml
   ```

### 问题：Headers 没有传递到上游

确保 Traefik 的 `authResponseHeaders` 包含所需的 headers：

```yaml
- "traefik.http.middlewares.auth.forwardauth.authResponseHeaders=X-Auth-User,X-Auth-Role,X-Auth-Method"
```

### 问题：环境变量未生效

1. 检查 `.env` 文件是否在正确位置
2. 确保配置中使用了 `env:VAR_NAME` 语法
3. 重启服务使环境变量生效

## 更多示例

查看主仓库的 `docs/` 目录获取更多配置示例和最佳实践。
