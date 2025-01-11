# AI-Powered i18n Translation System

基于 Go 语言开发的国际化翻译系统，支持多语言文档翻译、任务管理和监控。

## 功能特性

- 用户认证和授权（JWT）
- 文档翻译（支持 JSON 格式）
- 异步任务处理
- 任务状态监控
- 翻译结果下载
- 速率限制
- 性能监控（Prometheus）

## 环境要求

- Go 1.20+
- MongoDB
- Redis
- OpenAI API Key

## 快速开始

1. 克隆仓库

```bash
git clone https://github.com/yourusername/i18n-translation.git
cd i18n-translation
```

2. 安装依赖

```bash
go mod download
```

3. 配置环境

```bash
cp configs/apiserver.yaml.example configs/apiserver.yaml
# 编辑配置文件，填入必要的配置信息
```

4. 运行服务

```bash
go run cmd/i18n-apiserver/apiserver.go
```

## API 文档

### 认证相关

#### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 任务相关

#### 创建翻译任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_language": "en",
    "target_language": "zh",
    "content": {
      "greeting": "Hello, world!",
      "welcome": "Welcome to our platform"
    }
  }'
```

#### 执行翻译

```bash
curl -X POST http://localhost:8080/api/v1/tasks/TASK_ID/translate \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 获取任务状态

```bash
curl -X GET http://localhost:8080/api/v1/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 下载翻译结果

```bash
curl -X GET http://localhost:8080/api/v1/tasks/TASK_ID/download \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 监控

服务集成了 Prometheus 监控，可以通过 `/metrics` 端点获取监控指标。

## 配置说明

配置文件位于 `configs/apiserver.yaml`，主要配置项包括：

- 服务器配置（端口、模式等）
- 数据库配置（MongoDB 连接信息）
- Redis 配置
- JWT 配置
- 速率限制配置
- LLM API 配置

## 开发说明

### 目录结构

```
.
├── api/            # API 定义
├── cmd/            # 入口程序
├── configs/        # 配置文件
├── internal/       # 内部包
│   ├── apiserver/  # API 服务器
│   ├── pkg/        # 公共包
│   └── ...
├── pkg/            # 可导出的包
└── scripts/        # 脚本文件
```

### 测试

运行所有测试：

```bash
go test ./...
```

### 构建

```bash
make build
```

