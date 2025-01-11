# 构建阶段
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git make

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o i18n-apiserver cmd/i18n-apiserver/apiserver.go

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

# 复制二进制文件和配置
COPY --from=builder /app/i18n-apiserver /usr/local/bin/
COPY --from=builder /app/configs/apiserver.yaml /etc/i18n-translation/

# 使用非 root 用户运行
USER appuser

# 暴露端口
EXPOSE 8080

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/metrics || exit 1

# 运行应用
ENTRYPOINT ["i18n-apiserver", "--config=/etc/i18n-translation/apiserver.yaml"] 