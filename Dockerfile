# ==================== 构建阶段 ====================
FROM golang:1.22-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git ca-certificates

# 设置工作目录
WORKDIR /app

# 缓存依赖（利用 Docker 层缓存优化）
COPY go.mod go.sum* ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 Go 服务端
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/rpc-server ./cmd/rpc-server/

# ==================== 运行阶段 ====================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates bash

# 创建非 root 用户
RUN addgroup -S rpc-group && adduser -S rpc-user -G rpc-group

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /out/rpc-server /app/rpc-server

# 复制配置文件（如果有）
COPY configs/ /app/configs/

# 暴露端口（gRPC 默认 50051）
EXPOSE 50051

# 暴露健康检查端口
EXPOSE 8080

# 切换非 root 用户
USER rpc-user

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/rpc-server", "health-check"] || exit 1

# 启动入口
ENTRYPOINT ["/app/rpc-server"]
CMD ["start"]
