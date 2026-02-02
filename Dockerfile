# Dockerfile for manual builds
# For production releases, use Dockerfile.goreleaser

FROM golang:1.23-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o tiny-auth .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/tiny-auth .

# 复制示例配置（用户应挂载自己的配置）
COPY config.example.toml ./config.toml

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./tiny-auth", "server", "--config", "config.toml"]
