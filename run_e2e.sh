#!/bin/bash

# E2E 测试启动脚本
# 用法: ./run_e2e.sh

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

echo "=========================================="
echo "  AI 投资分析系统 - E2E 测试启动脚本"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查后端服务是否已运行
check_backend() {
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 检查前端服务是否已运行
check_frontend() {
    if curl -s http://localhost:5173 > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 启动后端服务
start_backend() {
    echo -e "${YELLOW}▶ 启动后端服务...${NC}"
    cd "$BACKEND_DIR"
    go run cmd/server/main.go &
    BACKEND_PID=$!
    echo -e "${GREEN}  ✓ 后端服务已启动 (PID: $BACKEND_PID)${NC}"

    # 等待后端服务就绪
    echo -e "${YELLOW}  等待后端服务就绪...${NC}"
    for i in {1..30}; do
        if check_backend; then
            echo -e "${GREEN}  ✓ 后端服务已就绪${NC}"
            return 0
        fi
        sleep 1
    done

    echo -e "${RED}  ✗ 后端服务启动超时${NC}"
    return 1
}

# 启动前端服务
start_frontend() {
    echo -e "${YELLOW}▶ 启动前端服务...${NC}"
    cd "$FRONTEND_DIR"
    npm run dev &
    FRONTEND_PID=$!
    echo -e "${GREEN}  ✓ 前端服务已启动 (PID: $FRONTEND_PID)${NC}"

    # 等待前端服务就绪
    echo -e "${YELLOW}  等待前端服务就绪...${NC}"
    for i in {1..30}; do
        if check_frontend; then
            echo -e "${GREEN}  ✓ 前端服务已就绪${NC}"
            return 0
        fi
        sleep 1
    done

    echo -e "${RED}  ✗ 前端服务启动超时${NC}"
    return 1
}

# 运行 E2E 测试
run_e2e_tests() {
    echo ""
    echo -e "${YELLOW}▶ 运行 E2E 测试...${NC}"
    cd "$FRONTEND_DIR"
    npm run test:e2e
    TEST_EXIT_CODE=$?

    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo ""
        echo -e "${GREEN}==========================================${NC}"
        echo -e "${GREEN}  ✓ E2E 测试全部通过！${NC}"
        echo -e "${GREEN}==========================================${NC}"
    else
        echo ""
        echo -e "${RED}==========================================${NC}"
        echo -e "${RED}  ✗ E2E 测试失败${NC}"
        echo -e "${RED}==========================================${NC}"
    fi

    return $TEST_EXIT_CODE
}

# 清理函数
cleanup() {
    echo ""
    echo -e "${YELLOW}▶ 清理进程...${NC}"

    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        echo -e "${GREEN}  ✓ 后端服务已停止${NC}"
    fi

    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
        echo -e "${GREEN}  ✓ 前端服务已停止${NC}"
    fi
}

# 注册清理函数
trap cleanup EXIT

# 主流程
main() {
    # 检查并启动后端
    if check_backend; then
        echo -e "${GREEN}✓ 后端服务已在运行${NC}"
    else
        start_backend
        if [ $? -ne 0 ]; then
            exit 1
        fi
    fi

    # 检查并启动前端
    if check_frontend; then
        echo -e "${GREEN}✓ 前端服务已在运行${NC}"
    else
        start_frontend
        if [ $? -ne 0 ]; then
            exit 1
        fi
    fi

    # 运行 E2E 测试
    run_e2e_tests
    exit $?
}

# 运行主流程
main
