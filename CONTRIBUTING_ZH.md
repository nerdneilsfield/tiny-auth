# Contributing to tiny-auth

[English](CONTRIBUTING.md) | 简体中文

感谢你考虑为 tiny-auth 做贡献！🎉

## 开发环境设置

### 前置要求

- Go 1.23 或更高版本
- [just](https://github.com/casey/just)（可选，也可用 make）
- [golangci-lint](https://golangci-lint.run/)
- [GoReleaser](https://goreleaser.com/)（仅发布时需要）

### 设置步骤

```bash
# 1. Fork 并克隆仓库
git clone https://github.com/YOUR_USERNAME/tiny-auth.git
cd tiny-auth

# 2. 安装依赖
just deps

# 3. 安装开发工具
just install-tools

# 4. 设置 Git hooks
just setup-hooks
```

## 开发流程

### 1. 创建功能分支

```bash
git checkout -b feature/your-feature-name
```

### 2. 进行更改

遵循以下规范：
- 使用中文注释代码
- 使用英文编写用户可见的消息
- 遵循 Go 代码规范
- 添加必要的测试

### 3. 运行测试和检查

```bash
# 格式化代码
just fmt

# 运行测试
just test

# 代码检查
just lint

# 完整检查
just check
```

### 4. 提交更改

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```bash
# 格式：<type>(<scope>): <subject>
git commit -m "feat(auth): add LDAP authentication support"
git commit -m "fix(config): resolve environment variable parsing issue"
git commit -m "docs: update README with new examples"
```

**类型（type）**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 添加测试
- `chore`: 构建/工具链更新

### 5. 推送并创建 Pull Request

```bash
git push origin feature/your-feature-name
```

然后在 GitHub 上创建 Pull Request。

## 代码规范

### Go 代码风格

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 和 `goimports` 格式化
- 通过 `golangci-lint` 检查
- 变量和函数使用驼峰命名
- 导出的标识符添加文档注释

### 注释规范

```go
// ✅ 好的注释（中文，解释为什么）
// TryBasic 尝试 Basic Auth 认证
// 使用常量时间比较防止时序攻击
func TryBasic(authHeader string, store *AuthStore) *AuthResult {
    // 解码 base64 凭证
    payload, err := base64.StdEncoding.DecodeString(...)
    ...
}

// ❌ 不好的注释（英文，只说做什么）
// Decode base64
payload, err := base64.StdEncoding.DecodeString(...)
```

### 安全规范

**必须**：
- 使用 `crypto/subtle.ConstantTimeCompare` 比较密码/token
- 清理所有 header 值（移除换行符）
- 不在日志中记录敏感信息（密码、token、API key）
- 验证所有用户输入

**推荐**：
- 限制字符串长度（防止 DoS）
- 添加超时控制
- 优雅处理错误

## 测试规范

### 单元测试

```go
// 文件名：xxx_test.go
package auth

import (
    "testing"
)

func TestTryBasic_Success(t *testing.T) {
    store := &AuthStore{
        BasicByUser: map[string]config.BasicAuthConfig{
            "admin": {Name: "admin", User: "admin", Pass: "secret", Roles: []string{"admin"}},
        },
    }
    
    result := TryBasic("Basic YWRtaW46c2VjcmV0", store)
    if result == nil {
        t.Fatal("expected success, got nil")
    }
    if result.User != "admin" {
        t.Errorf("expected user 'admin', got %q", result.User)
    }
}
```

### 集成测试

放在 `test/` 目录下。

### 测试覆盖率

- 目标：>80%
- 关键模块（auth、policy、config）：>90%

## 文档规范

### OpenSpec 规范

对于新功能，请先更新 OpenSpec 文档：

1. 在 `openspec/changes/` 下创建新目录
2. 编写 `proposal.md`（提案）
3. 创建 `specs/` 目录并编写详细规范
4. 编写 `design.md`（技术设计）
5. 编写 `tasks.md`（实现任务清单）

### README 更新

- 同时更新中英文版本（README.md 和 README_ZH.md）
- 保持两个版本内容同步
- 使用可折叠的 `<details>` 标签组织长内容
- 添加清晰的示例代码

## Pull Request 检查清单

在提交 PR 前，请确认：

- [ ] 代码通过所有测试（`just test`）
- [ ] 代码通过 lint 检查（`just lint`）
- [ ] 代码已格式化（`just fmt`）
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] PR 描述清晰说明了更改内容
- [ ] Commit 消息遵循 Conventional Commits 规范

## 发布流程（维护者）

1. 更新 `CHANGELOG.md`
2. 创建版本标签：`git tag v0.x.0`
3. 推送标签：`git push origin v0.x.0`
4. GitHub Actions 自动触发 GoReleaser
5. 检查 Release 页面和 Docker 镜像

## 获得帮助

- 💬 [GitHub Discussions](https://github.com/nerdneilsfield/tiny-auth/discussions) - 提问和讨论
- 🐛 [GitHub Issues](https://github.com/nerdneilsfield/tiny-auth/issues) - 报告 bug
- 📧 Email: dengqi935@gmail.com

## 行为准则

请友好、尊重地与他人互动。我们致力于提供一个开放和欢迎的环境。

---

再次感谢你的贡献！🙏
