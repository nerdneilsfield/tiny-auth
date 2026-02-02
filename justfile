# justfile for tiny-auth

# 默认任务：显示帮助
default:
    @just --list

# 变量
binary_name := "tiny-auth"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
date := `date -u +"%Y-%m-%dT%H:%M:%SZ"`
ldflags := "-s -w -X main.version=" + version + " -X main.buildTime=" + date + " -X main.gitCommit=" + commit

# 编译
build:
    @echo "Building {{binary_name}}..."
    go build -ldflags="{{ldflags}}" -o {{binary_name}} .
    @echo "✓ Build complete: {{binary_name}}"

# 编译并运行
run: build
    ./{{binary_name}} server

# 验证配置
validate:
    go run . validate config.toml

# 运行测试
test:
    go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# 运行测试并生成覆盖率报告
test-coverage: test
    go tool cover -html=coverage.txt -o coverage.html
    @echo "✓ Coverage report: coverage.html"

# 代码检查
lint:
    golangci-lint run ./...

# 代码格式化
fmt:
    gofmt -s -w .
    goimports -w .

# 清理构建产物
clean:
    rm -f {{binary_name}}
    rm -f coverage.txt coverage.html
    rm -rf dist/

# 安装依赖
deps:
    go mod download
    go mod tidy

# 安装开发工具
install-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/goreleaser/goreleaser/v2@latest

# 构建 Docker 镜像
docker-build:
    docker build -t {{binary_name}}:{{version}} .

# 运行 Docker Compose
docker-up:
    docker-compose up -d

# 停止 Docker Compose
docker-down:
    docker-compose down

# 查看 Docker 日志
docker-logs:
    docker-compose logs -f tiny-auth

# 创建发布（使用 GoReleaser）
release:
    goreleaser release --clean

# 快照发布（本地测试）
snapshot:
    goreleaser release --snapshot --clean

# 全部检查（测试 + lint）
check: test lint
    @echo "✓ All checks passed"

# 准备提交前检查
pre-commit: fmt check
    @echo "✓ Ready to commit"

# 安装预提交钩子
setup-hooks:
    @echo "Setting up git hooks..."
    @echo "#!/bin/sh" > .git/hooks/pre-commit
    @echo "just pre-commit" >> .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "✓ Git hooks installed"

# 显示版本信息
version:
    @echo "Version: {{version}}"
    @echo "Commit:  {{commit}}"
    @echo "Date:    {{date}}"
