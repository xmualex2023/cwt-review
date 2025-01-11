# API 文档

## 基础信息

- 基础路径：`http://localhost:8080/api/v1`
- 认证方式：Bearer Token
- 响应格式：JSON

## 错误码说明

| 错误码 | 说明             |
| ------ | ---------------- |
| 200    | 成功             |
| 400    | 请求参数错误     |
| 401    | 未认证或认证失败 |
| 403    | 权限不足         |
| 404    | 资源不存在       |
| 429    | 请求过于频繁     |
| 500    | 服务器内部错误   |

## 认证相关接口

### 1. 用户注册

**请求**

```http
POST /auth/register
Content-Type: application/json

{
    "username": "string",  // 用户名，3-32字符
    "password": "string"   // 密码，6-32字符
}
```

**测试命令**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**响应**

```json
{
  "code": 200,
  "message": "注册成功"
}
```

### 2. 用户登录

**请求**

```http
POST /auth/login
Content-Type: application/json

{
    "username": "string",
    "password": "string"
}
```

**测试命令**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**响应**

```json
{
  "code": 200,
  "data": {
    "token": "string",
    "expires_at": "2024-02-22T15:04:05Z",
    "refresh_after": "2024-02-21T15:04:05Z"
  }
}
```

### 3. 刷新令牌

**请求**

```http
POST /auth/refresh
Authorization: Bearer <token>
```

**测试命令**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Authorization: Bearer YOUR_OLD_TOKEN"
```

**响应**

```json
{
  "code": 200,
  "data": {
    "token": "string",
    "expires_at": "2024-02-22T15:04:05Z",
    "refresh_after": "2024-02-21T15:04:05Z"
  }
}
```

## 任务相关接口

### 1. 创建翻译任务

**请求**

```http
POST /tasks
Authorization: Bearer <token>
Content-Type: application/json

{
    "source_language": "string",  // 源语言代码，如 "en"
    "target_language": "string",  // 目标语言代码，如 "zh"
    "content": {                  // 需要翻译的内容
        "key1": "string",
        "key2": "string"
    }
}
```

**测试命令**

```bash
# 创建简单翻译任务
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

# 创建复杂翻译任务
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_language": "en",
    "target_language": "zh",
    "content": {
      "title": "Game Instructions",
      "sections": {
        "basic": "Basic Controls",
        "advanced": "Advanced Features",
        "tips": ["Tip 1", "Tip 2", "Tip 3"]
      },
      "messages": {
        "start": "Press Start to begin",
        "pause": "Game Paused",
        "end": "Game Over"
      }
    }
  }'
```

**响应**

```json
{
  "code": 200,
  "data": {
    "task_id": "string",
    "status": "created",
    "created_at": "2024-02-22T15:04:05Z"
  }
}
```

### 2. 执行翻译任务

**请求**

```http
POST /tasks/{task_id}/translate
Authorization: Bearer <token>
```

**测试命令**

```bash
curl -X POST http://localhost:8080/api/v1/tasks/TASK_ID/translate \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应**

```json
{
  "code": 200,
  "data": {
    "task_id": "string",
    "status": "processing"
  }
}
```

### 3. 获取任务状态

**请求**

```http
GET /tasks/{task_id}
Authorization: Bearer <token>
```

**测试命令**

```bash
curl -X GET http://localhost:8080/api/v1/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应**

```json
{
  "code": 200,
  "data": {
    "task_id": "string",
    "status": "string", // created/processing/completed/failed
    "progress": 0.75, // 翻译进度，0-1
    "created_at": "2024-02-22T15:04:05Z",
    "updated_at": "2024-02-22T15:04:05Z",
    "error": "string" // 如果失败，这里会有错误信息
  }
}
```

### 4. 下载翻译结果

**请求**

```http
GET /tasks/{task_id}/download
Authorization: Bearer <token>
```

**测试命令**

```bash
# 下载翻译结果
curl -X GET http://localhost:8080/api/v1/tasks/TASK_ID/download \
  -H "Authorization: Bearer YOUR_TOKEN"

# 下载并保存到文件
curl -X GET http://localhost:8080/api/v1/tasks/TASK_ID/download \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o translation_result.json
```

**响应**

```json
{
  "code": 200,
  "data": {
    "source_language": "en",
    "target_language": "zh",
    "original_content": {
      "key1": "string",
      "key2": "string"
    },
    "translated_content": {
      "key1": "string",
      "key2": "string"
    },
    "completed_at": "2024-02-22T15:04:05Z"
  }
}
```

## 完整测试流程示例

以下是一个完整的测试流程，从注册到获取翻译结果：

```bash
#!/bin/bash

# 1. 注册用户
echo "注册用户..."
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'

echo -e "\n"

# 2. 登录并获取令牌
echo "登录..."
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }' | jq -r '.data.token')

echo "获取到的令牌: $TOKEN"
echo -e "\n"

# 3. 创建翻译任务
echo "创建翻译任务..."
TASK_ID=$(curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_language": "en",
    "target_language": "zh",
    "content": {
      "title": "Game Instructions",
      "messages": {
        "start": "Press Start to begin",
        "pause": "Game Paused"
      }
    }
  }' | jq -r '.data.task_id')

echo "创建的任务 ID: $TASK_ID"
echo -e "\n"

# 4. 执行翻译
echo "执行翻译..."
curl -X POST "http://localhost:8080/api/v1/tasks/$TASK_ID/translate" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n"

# 5. 等待并检查状态
echo "检查任务状态..."
for i in {1..5}; do
  curl -X GET "http://localhost:8080/api/v1/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN"
  echo -e "\n"
  sleep 2
done

# 6. 下载结果
echo "下载翻译结果..."
curl -X GET "http://localhost:8080/api/v1/tasks/$TASK_ID/download" \
  -H "Authorization: Bearer $TOKEN" \
  -o translation_result.json

echo "翻译结果已保存到 translation_result.json"
```

## 监控接口

### 获取监控指标

**请求**

```http
GET /metrics
```

**测试命令**

```bash
curl -X GET http://localhost:8080/metrics
```

**响应**

```text
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/tasks"} 100
...
```
