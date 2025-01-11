#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 服务器地址
API_BASE="http://localhost:8080/api/v1"

# 测试数据
USERNAME="testuser_$(date +%s)"
PASSWORD="password123"
TOKEN=""
TASK_ID=""

# 辅助函数
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_response() {
    if [[ $1 == *"\"code\":200"* ]]; then
        log_success "$2"
        return 0
    else
        log_error "$3"
        echo "Response: $1"
        return 1
    fi
}

# 1. 测试用户注册
log_info "测试用户注册..."
RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$USERNAME\",
        \"password\": \"$PASSWORD\"
    }")

check_response "$RESPONSE" "用户注册成功" "用户注册失败" || exit 1

# 2. 测试用户登录
log_info "测试用户登录..."
RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$USERNAME\",
        \"password\": \"$PASSWORD\"
    }")

check_response "$RESPONSE" "用户登录成功" "用户登录失败" || exit 1
TOKEN=$(echo $RESPONSE | jq -r '.data.token')

# 3. 测试创建翻译任务
log_info "测试创建翻译任务..."
RESPONSE=$(curl -s -X POST "$API_BASE/tasks" \
    -H "Authorization: Bearer $TOKEN" \
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
    }')

check_response "$RESPONSE" "创建翻译任务成功" "创建翻译任务失败" || exit 1
TASK_ID=$(echo $RESPONSE | jq -r '.data.task_id')

# 4. 测试执行翻译
log_info "测试执行翻译..."
RESPONSE=$(curl -s -X POST "$API_BASE/tasks/$TASK_ID/translate" \
    -H "Authorization: Bearer $TOKEN")

check_response "$RESPONSE" "执行翻译成功" "执行翻译失败" || exit 1

# 5. 测试获取任务状态
log_info "测试获取任务状态..."
for i in {1..5}; do
    log_info "第 $i 次检查状态..."
    RESPONSE=$(curl -s -X GET "$API_BASE/tasks/$TASK_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    STATUS=$(echo $RESPONSE | jq -r '.data.status')
    if [[ "$STATUS" == "completed" ]]; then
        log_success "翻译完成"
        break
    elif [[ "$STATUS" == "failed" ]]; then
        log_error "翻译失败"
        exit 1
    else
        log_info "任务状态: $STATUS"
        sleep 2
    fi
done

# 6. 测试下载翻译结果
log_info "测试下载翻译结果..."
RESPONSE=$(curl -s -X GET "$API_BASE/tasks/$TASK_ID/download" \
    -H "Authorization: Bearer $TOKEN")

check_response "$RESPONSE" "下载翻译结果成功" "下载翻译结果失败" || exit 1

# 保存翻译结果
echo $RESPONSE | jq '.' > translation_result.json
log_success "翻译结果已保存到 translation_result.json"

# 7. 测试刷新令牌
log_info "测试刷新令牌..."
RESPONSE=$(curl -s -X POST "$API_BASE/auth/refresh" \
    -H "Authorization: Bearer $TOKEN")

check_response "$RESPONSE" "刷新令牌成功" "刷新令牌失败" || exit 1
NEW_TOKEN=$(echo $RESPONSE | jq -r '.data.token')

# 8. 测试速率限制
log_info "测试速率限制..."
for i in {1..110}; do
    RESPONSE=$(curl -s -X GET "$API_BASE/tasks/$TASK_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    if [[ "$RESPONSE" == *"429"* ]]; then
        log_success "速率限制测试成功"
        break
    fi
done

# 9. 测试监控指标
log_info "测试监控指标..."
RESPONSE=$(curl -s -X GET "http://localhost:8080/metrics")

if [[ -n "$RESPONSE" ]]; then
    log_success "获取监控指标成功"
else
    log_error "获取监控指标失败"
fi

log_success "所有测试完成" 