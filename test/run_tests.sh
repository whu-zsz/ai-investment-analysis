# 注意：本脚本默认不运行带 integration 标签、依赖 MySQL 的集成测试。
# 如需运行，请在 backend/ 目录下单独执行：go test -tags integration ./internal/repository/...

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
    echo -e "${CYAN}║           AI 投资分析系统 - 测试执行器                        ║${NC}"
    echo -e "${CYAN}║                   Test Runner                                ║${NC}"
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

# 打印帮助信息
print_help() {
    echo ""
    echo -e "${CYAN}用法:${NC}"
    echo -e "  ./test/run_tests.sh [选项]"
    echo ""
    echo -e "${CYAN}选项:${NC}"
    echo -e "  --with-integration    运行集成测试（需要本地 MySQL）"
    echo -e "  --with-benchmarks     运行性能基准测试"
    echo -e "  --help                显示帮助信息"
    echo ""
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

# 运行集成测试
run_integration_tests() {
    echo ""
    echo -e "${YELLOW}▶ 正在运行集成测试 (需要本地 MySQL)...${NC}"
    echo -e "${BLUE}  测试数据库: stock_analysis_test${NC}"
    print_separator

    local start=$(date +%s.%N)

    # 运行集成测试
    local output
    output=$(cd "$BACKEND_DIR" && go test -tags integration -v ./internal/repository/integration/... 2>&1)
    local exit_code=$?

    local end=$(date +%s.%N)
    local duration=$(echo "$end - $start" | bc | xargs printf "%.2f")

    # 解析测试结果
    local passed=0
    local failed=0
    local total=0

    while IFS= read -r line; do
        if [[ "$line" == *"=== RUN"* && "$line" != *"=== RUN   Test"*"_Integration/"* ]]; then
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

    echo ""
    if [ "$failed" -eq 0 ]; then
        echo -e "${GREEN}  ✓ 集成测试通过${NC}"
    else
        echo -e "${RED}  ✗ 集成测试失败 (${failed} 个失败)${NC}"
    fi
    echo -e "    通过: ${passed} | 失败: ${failed} | 耗时: ${duration}s"

    # 更新全局统计
    TOTAL_PASSED=$((TOTAL_PASSED + passed))
    TOTAL_FAILED=$((TOTAL_FAILED + failed))

    return $exit_code
}

# 运行基准测试
run_benchmark_tests() {
    echo ""
    echo -e "${YELLOW}▶ 正在运行性能基准测试...${NC}"
    print_separator

    local start=$(date +%s.%N)

    # 运行基准测试
    local output
    output=$(cd "$BACKEND_DIR" && go test ./internal/... -bench="Benchmark" -benchmem -count=1 2>&1)
    local exit_code=$?

    local end=$(date +%s.%N)
    local duration=$(echo "$end - $start" | bc | xargs printf "%.2f")

    # 解析并显示基准测试结果
    echo ""
    echo -e "${CYAN}  性能测试结果:${NC}"
    echo ""

    # 表头
    printf "  %-50s %-15s %-15s %-15s\n" "测试用例" "迭代次数" "每次耗时" "内存分配"
    echo -e "  ${BLUE}────────────────────────────────────────────────────────────────────────────────────────────────${NC}"

    # 解析输出
    while IFS= read -r line; do
        if [[ "$line" == Benchmark* ]]; then
            local name=$(echo "$line" | awk '{print $1}')
            local iterations=$(echo "$line" | awk '{print $2}')
            local ns_per_op=$(echo "$line" | awk '{print $3}')
            # Go benchmark format: name iterations ns/op B/op allocs/op
            # $4="ns/op"(unit) $5=B/op_value $6="B/op"(unit) $7=allocs_value $8="allocs/op"(unit)
            local b_per_op=$(echo "$line" | awk '{print $5}')
            local allocs_per_op=$(echo "$line" | awk '{print $7}')

            # 格式化耗时
            local time_str
            if (( $(echo "$ns_per_op > 1000000" | bc -l) )); then
                time_str=$(echo "scale=2; $ns_per_op / 1000000" | bc | xargs printf "%.2f ms")
            elif (( $(echo "$ns_per_op > 1000" | bc -l) )); then
                time_str=$(echo "scale=2; $ns_per_op / 1000" | bc | xargs printf "%.2f µs")
            else
                time_str=$(printf "%.2f ns" "$ns_per_op")
            fi

            # 格式化内存
            local mem_str
            if [ "$b_per_op" != "" ] && [ "$b_per_op" != "0" ]; then
                if (( $(echo "$b_per_op > 1048576" | bc -l) )); then
                    mem_str=$(echo "scale=2; $b_per_op / 1048576" | bc | xargs printf "%.2f MB")
                elif (( $(echo "$b_per_op > 1024" | bc -l) )); then
                    mem_str=$(echo "scale=2; $b_per_op / 1024" | bc | xargs printf "%.2f KB")
                else
                    mem_str=$(printf "%s B" "$b_per_op")
                fi
            else
                mem_str="0 B"
            fi

            printf "  %-50s %-15s %-15s %-15s\n" "$name" "$iterations" "$time_str" "$mem_str"
        fi
    done <<< "$output"

    echo ""
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}  ✓ 基准测试完成${NC}"
    else
        echo -e "${RED}  ✗ 基准测试失败${NC}"
    fi
    echo -e "    耗时: ${duration}s"
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
    # 显示帮助信息
    if [ "$1" == "--help" ] || [ "$2" == "--help" ]; then
        print_header
        print_help
        exit 0
    fi

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
        "internal/model"
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

    # 运行集成测试（可选，默认跳过）
    if [ "$1" == "--with-integration" ] || [ "$2" == "--with-integration" ]; then
        run_integration_tests
        print_separator
    else
        echo ""
        echo -e "${YELLOW}▶ 跳过集成测试 (使用 --with-integration 参数运行)${NC}"
    fi

    # 运行基准测试（可选，默认跳过）
    if [ "$1" == "--with-benchmarks" ] || [ "$2" == "--with-benchmarks" ]; then
        run_benchmark_tests
        print_separator
    else
        echo ""
        echo -e "${YELLOW}▶ 跳过基准测试 (使用 --with-benchmarks 参数运行)${NC}"
    fi

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
