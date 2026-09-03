# 多阶段构建：编译 + 运行
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 安装编译依赖
RUN apk add --no-cache git make

# 复制 go.mod 和 go.sum，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译 API Server 和 Worker
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/server/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o worker cmd/server/worker/main.go

# ============================================
# API Server 镜像
# ============================================
FROM alpine:latest AS api

WORKDIR /app

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 从 builder 复制编译好的二进制文件
COPY --from=builder /build/api /app/api
COPY --from=builder /build/config/config.docker.yaml /app/config/config.yaml

# 创建日志和上传目录
RUN mkdir -p /app/log /app/uploads

EXPOSE 8089

CMD ["/app/api"]

# ============================================
# Worker 镜像
# ============================================
FROM alpine:latest AS worker

WORKDIR /app

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 从 builder 复制编译好的二进制文件
COPY --from=builder /build/worker /app/worker
COPY --from=builder /build/config/config.docker.yaml /app/config/config.yaml

# 创建日志目录
RUN mkdir -p /app/log

CMD ["/app/worker"]
