FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git make gcc musl-dev

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o dbbak ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# 创建工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/dbbak .
COPY --from=builder /app/web ./web

# 创建数据目录
RUN mkdir -p /data/backups /data/db

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./dbbak"]
