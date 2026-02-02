# Makefile for tiny-auth

BINARY_NAME=tiny-auth
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-s -w -X main.version=$(VERSION) -X main.buildTime=$(DATE) -X main.gitCommit=$(COMMIT)

.PHONY: help build run test lint fmt clean deps docker-build docker-up docker-down release check

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译程序
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "✓ Build complete: $(BINARY_NAME)"

run: build ## 编译并运行
	./$(BINARY_NAME) server

validate: ## 验证配置文件
	go run . validate config.toml

test: ## 运行测试
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## 生成测试覆盖率报告
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✓ Coverage report: coverage.html"

lint: ## 代码检查
	golangci-lint run ./...

fmt: ## 代码格式化
	gofmt -s -w .
	goimports -w .

clean: ## 清理构建产物
	rm -f $(BINARY_NAME)
	rm -f coverage.txt coverage.html
	rm -rf dist/

deps: ## 安装依赖
	go mod download
	go mod tidy

install-tools: ## 安装开发工具
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/goreleaser/goreleaser/v2@latest

docker-build: ## 构建 Docker 镜像
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-up: ## 启动 Docker Compose
	docker-compose up -d

docker-down: ## 停止 Docker Compose
	docker-compose down

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f tiny-auth

release: ## 创建发布（使用 GoReleaser）
	goreleaser release --clean

snapshot: ## 快照发布（本地测试）
	goreleaser release --snapshot --clean

check: test lint ## 全部检查
	@echo "✓ All checks passed"

pre-commit: fmt check ## 提交前检查
	@echo "✓ Ready to commit"

.DEFAULT_GOAL := help
