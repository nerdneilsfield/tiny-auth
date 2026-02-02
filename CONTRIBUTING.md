# Contributing to tiny-auth

[简体中文](CONTRIBUTING_ZH.md) | English

Thank you for considering contributing to tiny-auth! 🎉

## Development Environment Setup

### Prerequisites

- Go 1.23 or higher
- [just](https://github.com/casey/just) (optional, can use make)
- [golangci-lint](https://golangci-lint.run/)
- [GoReleaser](https://goreleaser.com/) (only needed for releases)

### Setup Steps

```bash
# 1. Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/tiny-auth.git
cd tiny-auth

# 2. Install dependencies
just deps

# 3. Install development tools
just install-tools

# 4. Setup Git hooks
just setup-hooks
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Follow these guidelines:
- Use Chinese for code comments
- Use English for user-facing messages
- Follow Go coding standards
- Add necessary tests

### 3. Run Tests and Checks

```bash
# Format code
just fmt

# Run tests
just test

# Code linting
just lint

# Full check
just check
```

### 4. Commit Changes

We use [Conventional Commits](https://www.conventionalcommits.org/) specification:

```bash
# Format: <type>(<scope>): <subject>
git commit -m "feat(auth): add LDAP authentication support"
git commit -m "fix(config): resolve environment variable parsing issue"
git commit -m "docs: update README with new examples"
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation update
- `style`: Code formatting (no functional changes)
- `refactor`: Refactoring
- `perf`: Performance optimization
- `test`: Add tests
- `chore`: Build/toolchain updates

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Code Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Format with `gofmt` and `goimports`
- Pass `golangci-lint` checks
- Use camelCase for variables and functions
- Add documentation comments for exported identifiers

### Comment Guidelines

```go
// ✅ Good comment (Chinese, explains why)
// TryBasic 尝试 Basic Auth 认证
// 使用常量时间比较防止时序攻击
func TryBasic(authHeader string, store *AuthStore) *AuthResult {
    // 解码 base64 凭证
    payload, err := base64.StdEncoding.DecodeString(...)
    ...
}

// ❌ Bad comment (English, only describes what)
// Decode base64
payload, err := base64.StdEncoding.DecodeString(...)
```

### Security Guidelines

**Must**:
- Use `crypto/subtle.ConstantTimeCompare` for password/token comparison
- Sanitize all header values (remove newlines)
- Do not log sensitive information (passwords, tokens, API keys)
- Validate all user input

**Recommended**:
- Limit string length (prevent DoS)
- Add timeout controls
- Handle errors gracefully

## Testing Guidelines

### Unit Tests

```go
// Filename: xxx_test.go
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

### Integration Tests

Place in `test/` directory.

### Test Coverage

- Target: >80%
- Critical modules (auth, policy, config): >90%

## Documentation Guidelines

### OpenSpec Convention

For new features, update OpenSpec documents first:

1. Create new directory in `openspec/changes/`
2. Write `proposal.md` (proposal)
3. Create `specs/` directory with detailed specifications
4. Write `design.md` (technical design)
5. Write `tasks.md` (implementation task list)

### README Updates

- Update both English and Chinese versions (README.md and README_ZH.md)
- Keep both versions in sync
- Use collapsible `<details>` tags for long content
- Add clear code examples

## Pull Request Checklist

Before submitting PR, ensure:

- [ ] Code passes all tests (`just test`)
- [ ] Code passes lint checks (`just lint`)
- [ ] Code is formatted (`just fmt`)
- [ ] Added necessary tests
- [ ] Updated relevant documentation
- [ ] PR description clearly explains the changes
- [ ] Commit messages follow Conventional Commits

## Release Process (Maintainers)

1. Update `CHANGELOG.md`
2. Create version tag: `git tag v0.x.0`
3. Push tag: `git push origin v0.x.0`
4. GitHub Actions automatically triggers GoReleaser
5. Check Release page and Docker images

## Getting Help

- 💬 [GitHub Discussions](https://github.com/nerdneilsfield/tiny-auth/discussions) - Ask questions and discuss
- 🐛 [GitHub Issues](https://github.com/nerdneilsfield/tiny-auth/issues) - Report bugs
- 📧 Email: dengqi935@gmail.com

## Code of Conduct

Please interact with others in a friendly and respectful manner. We are committed to providing an open and welcoming environment.

---

Thank you again for your contribution! 🙏
