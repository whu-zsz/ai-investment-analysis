#!/bin/bash

# ============================================
# AI 投资分析系统 - 单元测试执行脚本
# ============================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"

# 测试结果统计
TOTAL_TESTS=0
TOTAL_PASSED=0
TOTAL_FAILED=0
START_TIME=0

# 打印标题
print_header() {
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║           AI 投资分析系统 - 单元测试执行器                    ║${NC}"
    echo -e "${CYAN}║                   Unit Test Runner                           ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# 打印分隔线
print_separator() {
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
}

# 打印模块标题
print_module() {
    local module=$1
    echo ""
    echo -e "${YELLOW}▶ 开始测试模块: ${module}${NC}"
    echo -e "${BLUE}  文件路径: ${module}/*_test.go${NC}"
}

# 打印测试进度
print_progress() {
    local current=$1
    local total=$2
    local test_name=$3
    local status=$4

    if [ "$status" == "PASS" ]; then
        echo -e "  ${GREEN}✓${NC} [$current/$total] ${test_name}"
    else
        echo -e "  ${RED}✗${NC} [$current/$total] ${test_name}"
    fi
}

# 打印模块结果
print_module_result() {
    local module=$1
    local passed=$2
    local failed=$3
    local duration=$4

    echo ""
    if [ "$failed" -eq 0 ]; then
        echo -e "${GREEN}  ✓ 模块测试通过: ${module}${NC}"
    else
        echo -e "${RED}  ✗ 模块测试失败: ${module} (${failed} 个失败)${NC}"
    fi
    echo -e "    通过: ${passed} | 失败: ${failed} | 耗时: ${duration}"
}

# 运行单个模块的测试
run_module_test() {
    local module=$1
    local module_path="$BACKEND_DIR/$module"

    print_module "$module"

    local start=$(date +%s.%N)

    # 运行测试并捕获输出
    local output
    output=$(cd "$module_path" && go test -v -count=1 2>&1)
    local exit_code=$?

    local end=$(date +%s.%N)
    local duration=$(echo "$end - $start" | bc | xargs printf "%.2f")

    # 解析测试结果
    local passed=0
    local failed=0
    local total=0

    while IFS= read -r line; do
        if [[ "$line" == *"=== RUN"* ]]; then
            total=$((total + 1))
        elif [[ "$line" == *"--- PASS"* ]]; then
            passed=$((passed + 1))
            local test_name=$(echo "$line" | sed 's/--- PASS: //' | sed 's/ (.*)//')
            print_progress "$passed" "$total" "$test_name" "PASS"
        elif [[ "$line" == *"--- FAIL"* ]]; then
            failed=$((failed + 1))
            local test_name=$(echo "$line" | sed 's/--- FAIL: //' | sed 's/ (.*)//')
            print_progress "$((passed + failed))" "$total" "$test_name" "FAIL"
        fi
    done <<< "$output"

    print_module_result "$module" "$passed" "$failed" "${duration}s"

    # 更新全局统计
    TOTAL_PASSED=$((TOTAL_PASSED + passed))
    TOTAL_FAILED=$((TOTAL_FAILED + failed))

    return $exit_code
}

# 运行覆盖率测试
run_coverage_test() {
    echo ""
    echo -e "${YELLOW}▶ 正在生成测试覆盖率报告...${NC}"

    cd "$BACKEND_DIR"

    # 生成覆盖率数据
    go test ./... -coverprofile=coverage.out -covermode=atomic 2>/dev/null

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}  ✓ 覆盖率数据已生成: coverage.out${NC}"

        # 显示各模块覆盖率
        echo ""
        echo -e "${CYAN}  模块覆盖率统计:${NC}"

        go tool cover -func=coverage.out 2>/dev/null | grep -E "^stock-analysis-backend" | while read -r line; do
            local pct=$(echo "$line" | awk '{print $NF}')
            local file=$(echo "$line" | awk '{print $1}')

            # 根据覆盖率显示不同颜色
            local pct_num=$(echo "$pct" | sed 's/%//')
            if (( $(echo "$pct_num >= 80" | bc -l) )); then
                echo -e "    ${GREEN}●${NC} ${file}: ${pct}"
            elif (( $(echo "$pct_num >= 50" | bc -l) )); then
                echo -e "    ${YELLOW}●${NC} ${file}: ${pct}"
            else
                echo -e "    ${RED}●${NC} ${file}: ${pct}"
            fi
        done

        # 生成 HTML 报告
        go tool cover -html=coverage.out -o coverage.html 2>/dev/null
        echo ""
        echo -e "${GREEN}  ✓ HTML 覆盖率报告已生成: backend/coverage.html${NC}"

        # 清理
        rm -f coverage.out
    fi
}

# 打印最终结果
print_final_result() {
    local end_time=$(date +%s.%N)
    local total_duration=$(echo "$end_time - $START_TIME" | bc | xargs printf "%.2f")

    TOTAL_TESTS=$((TOTAL_PASSED + TOTAL_FAILED))

    echo ""
    print_separator
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║                      测试执行完成                             ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # 结果统计表格
    echo -e "  ${BLUE}统计项目${NC}        ${BLUE}数值${NC}"
    echo -e "  ────────────────────────────"
    echo -e "  总测试用例:      ${TOTAL_TESTS}"
    echo -e "  通过用例:        ${GREEN}${TOTAL_PASSED}${NC}"
    echo -e "  失败用例:        ${RED}${TOTAL_FAILED}${NC}"
    echo -e "  通过率:          $(echo "scale=1; $TOTAL_PASSED * 100 / $TOTAL_TESTS" | bc)%"
    echo -e "  总耗时:          ${total_duration}s"
    echo ""

    # 最终状态
    if [ "$TOTAL_FAILED" -eq 0 ]; then
        echo -e "${GREEN}  ████████████████████████████████████████████████████████████${NC}"
        echo -e "${GREEN}  █                                                         █${NC}"
        echo -e "${GREEN}  █          🎉 所有测试通过！测试套件执行成功！ 🎉          █${NC}"
        echo -e "${GREEN}  █                                                         █${NC}"
        echo -e "${GREEN}  ████████████████████████████████████████████████████████████${NC}"
    else
        echo -e "${RED}  ████████████████████████████████████████████████████████████${NC}"
        echo -e "${RED}  █                                                         █${NC}"
        echo -e "${RED}  █          ⚠️ 存在测试失败，请检查错误日志！ ⚠️           █${NC}"
        echo -e "${RED}  █                                                         █${NC}"
        echo -e "${RED}  ████████████████████████████████████████████████████████████${NC}"
    fi
    echo ""
}

# 主函数
main() {
    START_TIME=$(date +%s.%N)

    print_header

    # 检查 Go 环境
    if ! command -v go &> /dev/null; then
        echo -e "${RED}错误: 未找到 Go 环境，请先安装 Go${NC}"
        exit 1
    fi

    # 检查后端目录
    if [ ! -d "$BACKEND_DIR" ]; then
        echo -e "${RED}错误: 未找到后端目录: $BACKEND_DIR${NC}"
        exit 1
    fi

    echo -e "${YELLOW}开始执行单元测试...${NC}"
    echo -e "${BLUE}Go 版本: $(go version | awk '{print $3}')${NC}"
    print_separator

    # 测试模块列表（按层级顺序）
    local modules=(
        "internal/utils"
        "internal/middleware"
        "internal/repository"
        "internal/service"
        "internal/handler"
    )

    # 运行各模块测试
    for module in "${modules[@]}"; do
        if [ -d "$BACKEND_DIR/$module" ]; then
            run_module_test "$module"
            print_separator
        fi
    done

    # 运行覆盖率测试
    run_coverage_test

    # 打印最终结果
    print_final_result

    # 返回退出码
    if [ "$TOTAL_FAILED" -gt 0 ]; then
        exit 1
    fi
    exit 0
}

# 执行主函数
main "$@"
