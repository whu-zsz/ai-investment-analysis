#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8080/api/v1"
TOKEN=""

echo "========================================="
echo "  Stock Analysis API 测试脚本"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local auth=$4

    echo -e "${GREEN}测试: $method $endpoint${NC}"

    if [ "$auth" = "auth" ]; then
        if [ -z "$TOKEN" ]; then
            echo -e "${RED}错误: 未登录，跳过测试${NC}"
            return
        fi
        auth_header="-H \"Authorization: Bearer $TOKEN\""
    else
        auth_header=""
    fi

    if [ -z "$data" ]; then
        cmd="curl -s -X $method $BASE_URL$endpoint $auth_header"
    else
        cmd="curl -s -X $method $BASE_URL$endpoint -H \"Content-Type: application/json\" $auth_header -d '$data'"
    fi

    response=$(eval $cmd)

    if echo "$response" | grep -q '"code":200'; then
        echo -e "${GREEN}✓ 成功${NC}"
    else
        echo -e "${RED}✗ 失败${NC}"
    fi

    echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    echo ""
    sleep 1
}

# 1. 测试健康检查
echo "1. 健康检查"
curl -s http://localhost:8080/health | python3 -m json.tool
echo ""
echo ""

# 2. 测试用户注册
echo "2. 用户注册"
test_api "POST" "/auth/register" '{"username":"testuser","email":"test@example.com","password":"123456"}'

# 3. 测试用户登录
echo "3. 用户登录"
response=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}')

TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}登录失败，无法继续测试${NC}"
    exit 1
else
    echo -e "${GREEN}登录成功，Token: ${TOKEN:0:20}...${NC}"
fi
echo ""

# 4. 测试获取用户信息
echo "4. 获取用户信息"
test_api "GET" "/user/profile" "" "auth"

# 5. 测试上传文件
echo "5. 上传测试文件"
echo -e "${GREEN}上传 test_data.csv${NC}"
response=$(curl -s -X POST $BASE_URL/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test_data.csv")

if echo "$response" | grep -q '"code":200'; then
    echo -e "${GREEN}✓ 上传成功${NC}"
else
    echo -e "${RED}✗ 上传失败${NC}"
fi
echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
echo ""
sleep 2

# 6. 测试获取交易记录
echo "6. 获取交易记录列表"
test_api "GET" "/transactions?page=1&page_size=10" "" "auth"

# 7. 测试获取交易统计
echo "7. 获取交易统计"
test_api "GET" "/transactions/stats" "" "auth"

# 8. 测试获取持仓
echo "8. 获取持仓列表"
test_api "GET" "/portfolios" "" "auth"

# 9. 测试创建交易记录
echo "9. 创建交易记录"
test_api "POST" "/transactions" '{
  "transaction_date":"2024-08-01",
  "transaction_type":"buy",
  "asset_type":"stock",
  "asset_code":"000001",
  "asset_name":"平安银行",
  "quantity":"100",
  "price_per_unit":"12.50",
  "commission":"2.50"
}' "auth"

# 10. 测试AI分析（如果配置了Deepseek API）
echo "10. AI分析（需要配置Deepseek API）"
# test_api "POST" "/analysis/summary?start_date=2024-01-01&end_date=2024-08-31" "" "auth"
echo -e "${GREEN}跳过（需要配置Deepseek API）${NC}"
echo ""

# 11. 测试更新用户信息
echo "11. 更新用户信息"
test_api "PUT" "/user/profile" '{"investment_preference":"aggressive"}' "auth"

# 12. 测试获取上传历史
echo "12. 获取上传历史"
test_api "GET" "/upload/history" "" "auth"

echo "========================================="
echo -e "${GREEN}  测试完成！${NC}"
echo "========================================="
