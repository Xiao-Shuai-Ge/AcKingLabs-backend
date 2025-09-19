# 构建阶段
FROM golang:1.23 AS builder

# 设置环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://goproxy.io,direct

# 设置工作目录
WORKDIR /app

# 先单独复制依赖文件（重要优化）
COPY go.mod go.sum ./

# 下载依赖（利用Docker缓存层）
RUN go mod download

# 再复制项目代码（避免覆盖go.mod）
COPY . .

# 构建可执行文件（添加构建参数）
RUN go build -ldflags="-s -w" -o acking ./cmd/main.go

# ----------------------------
# 运行阶段
FROM alpine:latest

# 设置时区
RUN apk add --no-cache tzdata && \
    ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 设置工作目录
WORKDIR /app

# 从构建阶段复制产物
COPY --from=builder /app/acking ./
COPY --from=builder /app/config.yaml ./
# 复制静态资源目录
COPY --from=builder /app/static ./static

# 暴露端口
EXPOSE 8080

# 入口点
ENTRYPOINT ["./acking"]