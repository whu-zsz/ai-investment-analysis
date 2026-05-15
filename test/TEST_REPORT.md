# 后端单元测试报告

> **项目**: AI 投资分析系统后端
> **执行日期**: 2026-04-24 ~ 2026-05-15
> **执行人**: 林润民（测试 & 文档）
> **Go 版本**: 1.26.1
> **测试状态**: ✅ **全部通过**

---

## 1. 测试概况

| 模块 | 测试文件 | 用例数 | 通过 | 失败 | 覆盖率 |
|------|----------|--------|------|------|--------|
| model/portfolio | `internal/model/portfolio_test.go` | 10 | 10 | 0 | - |
| model/tablename | `internal/model/table_name_test.go` | 8 | 8 | 0 | - |
| utils/crypto | `internal/utils/crypto_test.go` | 8 | 8 | 0 | 92.9% |
| utils/jwt | `internal/utils/jwt_test.go` | 10 | 10 | 0 | 92.9% |
| middleware/auth | `internal/middleware/auth_test.go` | 6 | 6 | 0 | 69.0% |
| handler/user | `internal/handler/user_test.go` | 8 | 8 | 0 | 86.9% |
| handler/transaction | `internal/handler/transaction_test.go` | 13 | 13 | 0 | 86.9% |
| handler/upload | `internal/handler/upload_test.go` | 8 | 8 | 0 | 86.9% |
| handler/portfolio | `internal/handler/portfolio_test.go` | 5 | 5 | 0 | 86.9% |
| handler/market | `internal/handler/market_test.go` | 10 | 10 | 0 | 86.9% |
| handler/analysis | `internal/handler/analysis_test.go` | 18 | 18 | 0 | 86.9% |
| service/user | `internal/service/user_service_test.go` | 11 | 11 | 0 | 76.4% |
| service/transaction | `internal/service/transaction_service_test.go` | 18 | 18 | 0 | 76.4% |
| service/ai | `internal/service/ai_service_test.go` | 19 | 19 | 0 | 76.4% |
| service/upload | `internal/service/upload_service_test.go` | 11 | 11 | 0 | 76.4% |
| service/portfolio | `internal/service/portfolio_service_test.go` | 14 | 14 | 0 | 76.4% |
| service/file_parser | `internal/service/file_parser_test.go` | 9 | 9 | 0 | 76.4% |
| service/market_snapshot | `internal/service/market_snapshot_service_test.go` | 8 | 8 | 0 | 76.4% |
| service/decimal_helpers | `internal/service/decimal_helpers_test.go` | 5 | 5 | 0 | 76.4% |
| service/market_data | `internal/service/market_data_service_test.go` | 10 | 10 | 0 | 76.4% |
| service/market_scheduler | `internal/service/market_scheduler_test.go` | 7 | 7 | 0 | 76.4% |
| service/stock_analysis_metric | `internal/service/stock_analysis_metric_service_test.go` | 10 | 10 | 0 | 76.4% |
| repository/user | `internal/repository/user_repo_test.go` | 21 | 21 | 0 | 34.2% |
| repository/transaction | `internal/repository/transaction_repo_test.go` | 25 | 25 | 0 | 34.2% |
| repository/portfolio | `internal/repository/portfolio_repo_test.go` | 21 | 21 | 0 | 34.2% |
| repository/uploaded_file | `internal/repository/uploaded_file_repo_test.go` | 13 | 13 | 0 | 34.2% |
| repository/analysis_task | `internal/repository/analysis_task_repo_test.go` | 21 | 21 | 0 | 34.2% |
| repository/analysis_report | `internal/repository/analysis_report_repo_test.go` | 22 | 22 | 0 | 34.2% |
| repository/analysis_report_item | `internal/repository/analysis_report_item_repo_test.go` | 8 | 8 | 0 | 34.2% |
| repository/market_snapshot | `internal/repository/market_snapshot_repo_test.go` | 20 | 20 | 0 | 34.2% |
| repository/stock_analysis_metric | `internal/repository/stock_analysis_metric_repo_test.go` | 17 | 17 | 0 | 34.2% |
| **单元测试总计** | **31 个文件** | **414** | **414** | **0** | **~72%** |
| integration/user | `internal/repository/integration/user_repo_integration_test.go` | 12 | 12 | 0 | - |
| integration/transaction | `internal/repository/integration/transaction_repo_integration_test.go` | 11 | 11 | 0 | - |
| integration/portfolio | `internal/repository/integration/portfolio_repo_integration_test.go` | 10 | 10 | 0 | - |
| **集成测试总计** | **3 个文件** | **33** | **33** | **0** | **-** |
| **总计** | **34 个文件** | **447** | **447** | **0** | **~72%** |

### 测试状态

✅ **全部通过** - 447 个测试用例，0 个失败

---

## 2. 本次新增测试模块 ⭐

### 2.0 integration/user_repo_integration_test.go (用户仓储集成测试) - 2026-05-06 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestUserRepository_Integration/Create | 创建用户 | ✅ PASS |
| TestUserRepository_Integration/FindByID | 按ID查找 | ✅ PASS |
| TestUserRepository_Integration/FindByID_NotFound | ID不存在 | ✅ PASS |
| TestUserRepository_Integration/FindByUsername | 按用户名查找 | ✅ PASS |
| TestUserRepository_Integration/FindByUsername_NotFound | 用户名不存在 | ✅ PASS |
| TestUserRepository_Integration/FindByEmail | 按邮箱查找 | ✅ PASS |
| TestUserRepository_Integration/FindByEmail_NotFound | 邮箱不存在 | ✅ PASS |
| TestUserRepository_Integration/Update | 更新用户 | ✅ PASS |
| TestUserRepository_Integration/Delete | 删除用户 | ✅ PASS |
| TestUserRepository_Integration/UpdateLastLogin | 更新登录时间 | ✅ PASS |
| TestUserRepository_Integration/UpdateTotalProfit | 更新总收益 | ✅ PASS |

### 2.0.1 integration/transaction_repo_integration_test.go (交易仓储集成测试) - 2026-05-06 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestTransactionRepository_Integration/Create | 创建交易 | ✅ PASS |
| TestTransactionRepository_Integration/BatchCreate | 批量创建 | ✅ PASS |
| TestTransactionRepository_Integration/FindByID | 按ID查找 | ✅ PASS |
| TestTransactionRepository_Integration/FindByID_NotFound | ID不存在 | ✅ PASS |
| TestTransactionRepository_Integration/FindByUserID | 按用户ID分页查找 | ✅ PASS |
| TestTransactionRepository_Integration/FindByAssetCode | 按资产代码查找 | ✅ PASS |
| TestTransactionRepository_Integration/FindByDateRange | 按日期范围查找 | ✅ PASS |
| TestTransactionRepository_Integration/Update | 更新交易 | ✅ PASS |
| TestTransactionRepository_Integration/Delete | 删除交易 | ✅ PASS |
| TestTransactionRepository_Integration/GetTransactionStats | 获取交易统计 | ✅ PASS |

### 2.0.2 integration/portfolio_repo_integration_test.go (持仓仓储集成测试) - 2026-05-06 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestPortfolioRepository_Integration/Create | 创建持仓 | ✅ PASS |
| TestPortfolioRepository_Integration/FindByID | 按ID查找 | ✅ PASS |
| TestPortfolioRepository_Integration/FindByID_NotFound | ID不存在 | ✅ PASS |
| TestPortfolioRepository_Integration/FindByUserID | 按用户ID查找 | ✅ PASS |
| TestPortfolioRepository_Integration/FindByUserAndAsset | 按用户和资产查找 | ✅ PASS |
| TestPortfolioRepository_Integration/FindByUserAndAsset_NotFound | 用户资产不存在 | ✅ PASS |
| TestPortfolioRepository_Integration/Update | 更新持仓 | ✅ PASS |
| TestPortfolioRepository_Integration/Delete | 删除持仓 | ✅ PASS |
| TestPortfolioRepository_Integration/UpdateCurrentPrice | 更新当前价格 | ✅ PASS |

### 2.1 model/portfolio_test.go (Portfolio 模型测试) - 2026-05-05 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestPortfolio_TableName | 表名验证 | ✅ PASS |
| TestPortfolio_BeforeSave_CalculatesMarketValue | 市值计算 | ✅ PASS |
| TestPortfolio_BeforeSave_CalculatesProfitLoss | 盈亏计算 | ✅ PASS |
| TestPortfolio_BeforeSave_CalculatesProfitLossPercent | 盈亏百分比计算 | ✅ PASS |
| TestPortfolio_BeforeSave_ZeroAverageCost | 成本为零避免除以零 | ✅ PASS |
| TestPortfolio_BeforeSave_NilCurrentPrice | 价格为 nil 不计算 | ✅ PASS |
| TestPortfolio_BeforeSave_NegativeProfitLoss | 亏损情况测试 | ✅ PASS |
| TestPortfolio_BeforeSave_DecimalPrecision | 小数精度计算 | ✅ PASS |
| TestPortfolio_BeforeSave_ReturnsNil | 返回 nil | ✅ PASS |
| TestPortfolio_BeforeSave_WithGormDB | GORM DB 参数测试 | ✅ PASS |

### 2.0.1 model/table_name_test.go (表名验证测试) - 2026-05-05 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestUser_TableName | users | ✅ PASS |
| TestTransaction_TableName | transactions | ✅ PASS |
| TestPortfolio_TableName | portfolios | ✅ PASS |
| TestMarketSnapshot_TableName | market_snapshots | ✅ PASS |
| TestAnalysisTask_TableName | ai_analysis_tasks | ✅ PASS |
| TestAnalysisReport_TableName | ai_analysis_reports | ✅ PASS |
| TestAnalysisReportItem_TableName | ai_analysis_report_items | ✅ PASS |
| TestStockAnalysisMetric_TableName | stock_analysis_metrics | ✅ PASS |
| TestUploadedFile_TableName | uploaded_files | ✅ PASS |

### 2.1 repository/market_snapshot_repo_test.go (市场快照仓储测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestMarketSnapshotRepository_BatchCreate | 批量创建快照 | ✅ PASS |
| TestMarketSnapshotRepository_BatchCreate_Empty | 空批量创建 | ✅ PASS |
| TestMarketSnapshotRepository_FindLatestBatchNo | 查找最新批次号 | ✅ PASS |
| TestMarketSnapshotRepository_FindLatestBatchNo_Empty | 空仓储查找 | ✅ PASS |
| TestMarketSnapshotRepository_FindByBatchNo | 按批次号查找 | ✅ PASS |
| TestMarketSnapshotRepository_FindLatestBySymbol | 按代码查找最新 | ✅ PASS |
| TestMarketSnapshotRepository_FindLatestBySymbol_NotFound | 找不到快照 | ✅ PASS |
| TestMarketSnapshotRepository_FindHistory | 查找历史 | ✅ PASS |
| TestMarketSnapshotRepository_FindHistoryBySymbol | 按代码查找历史 | ✅ PASS |
| TestMarketSnapshotRepository_Interface | 接口实现验证 | ✅ PASS |

### 2.2 repository/analysis_report_item_repo_test.go (报告项仓储测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestAnalysisReportItemRepository_BatchCreate | 批量创建报告项 | ✅ PASS |
| TestAnalysisReportItemRepository_BatchCreate_Empty | 空批量创建 | ✅ PASS |
| TestAnalysisReportItemRepository_FindByReportID | 按报告ID查找 | ✅ PASS |
| TestAnalysisReportItemRepository_FindByReportID_Empty | 空结果查找 | ✅ PASS |
| TestAnalysisReportItemRepository_Interface | 接口实现验证 | ✅ PASS |

### 2.3 repository/stock_analysis_metric_repo_test.go (股票分析指标仓储测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestStockAnalysisMetricRepository_Upsert | 创建指标 | ✅ PASS |
| TestStockAnalysisMetricRepository_Upsert_Update | 更新指标 | ✅ PASS |
| TestStockAnalysisMetricRepository_BatchUpsert | 批量创建 | ✅ PASS |
| TestStockAnalysisMetricRepository_BatchUpsert_Empty | 空批量创建 | ✅ PASS |
| TestStockAnalysisMetricRepository_FindByUserPeriod | 按用户和时间查找 | ✅ PASS |
| TestStockAnalysisMetricRepository_FindByUserPeriod_WithSymbols | 按股票代码过滤 | ✅ PASS |
| TestStockAnalysisMetricRepository_FindByUserPeriod_Empty | 空结果查找 | ✅ PASS |
| TestStockAnalysisMetricRepository_FindByUserSymbolPeriod | 按用户股票时间查找 | ✅ PASS |
| TestStockAnalysisMetricRepository_FindByUserSymbolPeriod_NotFound | 找不到指标 | ✅ PASS |
| TestStockAnalysisMetricRepository_Interface | 接口实现验证 | ✅ PASS |

### 2.4 service/market_snapshot_service_test.go (市场快照服务测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestMarketSnapshotService_GetLatestSnapshots | 获取最新快照 | ✅ PASS |
| TestMarketSnapshotService_GetLatestSnapshots_Empty | 空快照 | ✅ PASS |
| TestMarketSnapshotService_GetHistory | 获取历史 | ✅ PASS |
| TestMarketSnapshotService_GetHistory_BySymbol | 按代码获取历史 | ✅ PASS |
| TestMarketSnapshotService_GetDashboardSnapshot | 获取仪表盘快照 | ✅ PASS |
| TestMarketSnapshotService_GetDashboardSnapshot_Empty | 空仪表盘 | ✅ PASS |
| TestMarketSnapshotService_GetDashboardSnapshot_Stats | 统计计算 | ✅ PASS |
| TestMarketSnapshotService_Interface | 接口实现验证 | ✅ PASS |

### 2.5 service/decimal_helpers_test.go (小数辅助函数测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestModelDecimalZero | 获取零值 | ✅ PASS |
| TestModelDecimalFromInt | 从整数创建 | ✅ PASS |
| TestModelDecimalFromInt_LargeValue | 大整数测试 | ✅ PASS |
| TestModelDecimalFromInt_Arithmetic | 算术运算 | ✅ PASS |
| TestModelDecimalZero_Comparisons | 比较操作 | ✅ PASS |

### 2.6 service/market_data_service_test.go (市场数据服务测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestNormalizeSymbol | 股票代码标准化 | ✅ PASS |
| TestNormalizeSymbols_Multiple | 多股票代码标准化 | ✅ PASS |
| TestMarketDataService_FetchAndStoreQuotesBySymbols | 按代码获取行情 | ✅ PASS |
| TestMarketDataService_FetchAndStoreQuotesBySymbols_Empty | 空代码列表 | ✅ PASS |
| TestMarketDataService_FetchAndStoreQuotesBySymbols_ProviderError | 提供者错误 | ✅ PASS |
| TestMarketDataService_FetchAndStoreQuotesBySymbols_NoQuotes | 无行情返回 | ✅ PASS |
| TestMarketDataService_FetchAndStoreMarketSnapshots | 获取市场快照 | ✅ PASS |
| TestMarketDataService_FetchAndStoreMarketSnapshots_EmptySymbols | 空配置 | ✅ PASS |
| TestMarketDataService_BatchNo | 批次号生成 | ✅ PASS |
| TestMarketDataService_Interface | 接口实现验证 | ✅ PASS |

### 2.7 service/market_scheduler_test.go (市场调度器测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestNewMarketScheduler | 创建调度器 | ✅ PASS |
| TestNewMarketScheduler_DefaultInterval | 默认间隔 | ✅ PASS |
| TestMarketScheduler_Start | 启动调度器 | ✅ PASS |
| TestMarketScheduler_ContextCancellation | 上下文取消 | ✅ PASS |
| TestMarketScheduler_RunOnce | 单次运行 | ✅ PASS |
| TestMarketScheduler_RunOnce_Error | 单次运行错误 | ✅ PASS |
| TestMarketScheduler_Interface | 接口实现验证 | ✅ PASS |

### 2.8 service/stock_analysis_metric_service_test.go (股票分析指标服务测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestNormalizeSymbols | 股票代码标准化 | ✅ PASS |
| TestAggregateMetricTransactions | 交易聚合 | ✅ PASS |
| TestAggregateMetricTransactions_Empty | 空交易列表 | ✅ PASS |
| TestApplyMetricMarketHistory | 应用市场历史数据 | ✅ PASS |
| TestApplyMetricMarketHistory_HighLowPrice | 最高最低价计算 | ✅ PASS |
| TestStockAnalysisMetricService_PrepareMetrics | 准备指标 | ✅ PASS |
| TestStockAnalysisMetricService_PrepareMetrics_Empty | 空交易 | ✅ PASS |
| TestStockAnalysisMetricService_PrepareMetrics_Cached | 缓存测试 | ✅ PASS |
| TestStockAnalysisMetricService_Interface | 接口实现验证 | ✅ PASS |

---

## 2.1 Repository 边界条件测试 (2026-05-15 新增) ⭐

### user_repo_test.go (用户仓储边界测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestUserRepository_Boundary_Create_EmptyUsername | 空用户名创建 | ✅ PASS |
| TestUserRepository_Boundary_Create_EmptyEmail | 空邮箱创建 | ✅ PASS |
| TestUserRepository_Boundary_Create_EmptyPasswordHash | 空密码哈希创建 | ✅ PASS |
| TestUserRepository_Boundary_Create_LongStrings | 超长字符串创建 | ✅ PASS |
| TestUserRepository_Boundary_FindByID_Zero | ID=0 查找 | ✅ PASS |
| TestUserRepository_Boundary_FindByID_NotFound | 不存在的 ID 查找 | ✅ PASS |
| TestUserRepository_Boundary_FindByUsername_Empty | 空用户名查找 | ✅ PASS |
| TestUserRepository_Boundary_FindByUsername_NotFound | 不存在的用户名查找 | ✅ PASS |
| TestUserRepository_Boundary_FindByEmail_Empty | 空邮箱查找 | ✅ PASS |
| TestUserRepository_Boundary_FindByEmail_NotFound | 不存在的邮箱查找 | ✅ PASS |
| TestUserRepository_Boundary_Delete_Zero | ID=0 删除 | ✅ PASS |
| TestUserRepository_Boundary_Delete_NotFound | 不存在的 ID 删除 | ✅ PASS |

### transaction_repo_test.go (交易仓储边界测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestTransactionRepository_Boundary_Create_ZeroUserID | UserID=0 创建 | ✅ PASS |
| TestTransactionRepository_Boundary_Create_NegativeQuantity | 负数量创建 | ✅ PASS |
| TestTransactionRepository_Boundary_Create_ZeroPrice | 零价格创建 | ✅ PASS |
| TestTransactionRepository_Boundary_Create_EmptyAssetCode | 空资产代码创建 | ✅ PASS |
| TestTransactionRepository_Boundary_BatchCreate_EmptySlice | 空切片批量创建 | ✅ PASS |
| TestTransactionRepository_Boundary_BatchCreate_SingleItem | 单条记录批量创建 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByID_Zero | ID=0 查找 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByID_NotFound | 不存在的 ID 查找 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByUserID_ZeroLimit | limit=0 分页 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByUserID_LargeOffset | 大偏移量分页 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByAssetCode_EmptyCode | 空资产代码查找 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByDateRange_SameDay | 同一天日期范围 | ✅ PASS |
| TestTransactionRepository_Boundary_FindByDateRange_EmptyDates | 空日期范围 | ✅ PASS |
| TestTransactionRepository_Boundary_GetTransactionStats_NoTransactions | 无交易统计 | ✅ PASS |

### portfolio_repo_test.go (持仓仓储边界测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestPortfolioRepository_Boundary_Create_ZeroUserID | UserID=0 创建 | ✅ PASS |
| TestPortfolioRepository_Boundary_Create_EmptyAssetCode | 空资产代码创建 | ✅ PASS |
| TestPortfolioRepository_Boundary_Create_NegativeQuantity | 负数量创建 | ✅ PASS |
| TestPortfolioRepository_Boundary_FindByID_Zero | ID=0 查找 | ✅ PASS |
| TestPortfolioRepository_Boundary_FindByID_NotFound | 不存在的 ID 查找 | ✅ PASS |
| TestPortfolioRepository_Boundary_FindByUserID_NoPortfolios | 用户无持仓 | ✅ PASS |
| TestPortfolioRepository_Boundary_FindByUserAndAsset_EmptyCode | 空资产代码组合查找 | ✅ PASS |
| TestPortfolioRepository_Boundary_FindByUserAndAsset_NotFound | 不存在的组合查找 | ✅ PASS |
| TestPortfolioRepository_Boundary_UpdateCurrentPrice_EmptyCode | 空代码更新价格 | ✅ PASS |
| TestPortfolioRepository_Boundary_UpdateCurrentPrice_NegativePrice | 负价格更新 | ✅ PASS |
| TestPortfolioRepository_Boundary_UpdateCurrentPrice_ZeroPrice | 零价格更新 | ✅ PASS |
| TestPortfolioRepository_Boundary_Delete_Zero | ID=0 删除 | ✅ PASS |

### 其他 Repository 边界测试

| 测试文件 | 新增用例 | 结果 |
|----------|---------|------|
| uploaded_file_repo_test.go | 6 | ✅ 全部通过 |
| analysis_task_repo_test.go | 8 | ✅ 全部通过 |
| analysis_report_repo_test.go | 10 | ✅ 全部通过 |
| analysis_report_item_repo_test.go | 3 | ✅ 全部通过 |
| market_snapshot_repo_test.go | 10 | ✅ 全部通过 |
| stock_analysis_metric_repo_test.go | 7 | ✅ 全部通过 |

---

## 3. 覆盖率对比

### 本次更新前后对比

| 模块 | 更新前 | 更新后 | 变化 |
|------|--------|--------|------|
| model | 0% | - | 新增 |
| handler | 86.9% | 86.9% | - |
| service | 52.7% | 76.4% | +23.7% |
| repository | 31.9% | 34.2% | +2.3% ⭐ |
| repository integration | 0% | - | 新增 (真实数据库测试) |
| middleware | 69.0% | 69.0% | - |
| utils | 92.9% | 92.9% | - |

### 本次新增测试文件

| 文件 | 新增用例数 |
|------|-----------|
| portfolio_test.go | 10 |
| table_name_test.go | 8 |
| market_snapshot_repo_test.go | 10 |
| analysis_report_item_repo_test.go | 5 |
| stock_analysis_metric_repo_test.go | 10 |
| market_snapshot_service_test.go | 8 |
| decimal_helpers_test.go | 5 |
| market_data_service_test.go | 10 |
| market_scheduler_test.go | 7 |
| stock_analysis_metric_service_test.go | 10 |
| integration/test_helpers_integration.go | - |
| integration/user_repo_integration_test.go | 12 |
| integration/transaction_repo_integration_test.go | 11 |
| integration/portfolio_repo_integration_test.go | 10 |
| **边界条件测试合计** | **82** ⭐ |
| **总计** | **198** |

---

## 4. 运行命令

### 4.1 基本测试命令

```bash
# 进入后端目录
cd /Users/lnm/Downloads/stock_whu/ai-investment-analysis/backend

# 运行所有单元测试
go test ./internal/... -v

# 查看覆盖率
go test ./internal/... -cover

# 生成覆盖率报告
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 4.2 运行特定模块测试

```bash
# Repository 测试
go test ./internal/repository/... -v

# Service 测试
go test ./internal/service/... -v

# Handler 测试
go test ./internal/handler/... -v
```

---

## 5. 测试文件位置

```
backend/
├── internal/
│   ├── model/
│   │   ├── portfolio_test.go               ⭐ 新增
│   │   └── table_name_test.go              ⭐ 新增
│   ├── utils/
│   │   ├── crypto_test.go
│   │   └── jwt_test.go
│   ├── middleware/
│   │   └── auth_test.go
│   ├── handler/
│   │   ├── user_test.go
│   │   ├── transaction_test.go
│   │   ├── upload_test.go
│   │   ├── portfolio_test.go
│   │   ├── market_test.go
│   │   └── analysis_test.go
│   ├── service/
│   │   ├── user_service_test.go
│   │   ├── transaction_service_test.go
│   │   ├── upload_service_test.go
│   │   ├── portfolio_service_test.go
│   │   ├── ai_service_test.go
│   │   ├── file_parser_test.go
│   │   ├── market_snapshot_service_test.go         ⭐ 新增
│   │   ├── decimal_helpers_test.go                  ⭐ 新增
│   │   ├── market_data_service_test.go              ⭐ 新增
│   │   ├── market_scheduler_test.go                 ⭐ 新增
│   │   └── stock_analysis_metric_service_test.go    ⭐ 新增
│   └── repository/
│       ├── user_repo_test.go
│       ├── transaction_repo_test.go
│       ├── portfolio_repo_test.go
│       ├── uploaded_file_repo_test.go
│       ├── analysis_task_repo_test.go
│       ├── analysis_report_repo_test.go
│       ├── analysis_report_item_repo_test.go        ⭐ 新增
│       ├── market_snapshot_repo_test.go             ⭐ 新增
│       ├── stock_analysis_metric_repo_test.go       ⭐ 新增
│       └── integration/                             ⭐ 新增 (集成测试)
│           ├── test_helpers_integration.go
│           ├── user_repo_integration_test.go
│           ├── transaction_repo_integration_test.go
│           └── portfolio_repo_integration_test.go
```

---

## 6. 测试完成情况

### 6.1 已完成 ✅

- [x] Handler 层全部模块测试
- [x] Service 层全部核心模块测试
- [x] Repository 层全部模块测试
- [x] Repository 层边界条件测试 ⭐
- [x] Middleware 层测试
- [x] Utils 层测试

### 6.2 测试覆盖统计

| 层级 | 测试文件数 | 测试用例数 | 覆盖率 |
|------|-----------|-----------|--------|
| Model | 2 | 18 | - |
| Handler | 6 | 72 | 86.9% |
| Service | 11 | 120 | 76.4% |
| Repository | 9 | 169 | 34.2% |
| Repository Integration | 3 | 33 | - |
| Middleware | 1 | 13 | 69.0% |
| Utils | 2 | 22 | 92.9% |
| **总计** | **34** | **447** | **~72%** |

### 6.3 测试运行命令

#### 单元测试

```bash
# 运行所有单元测试（推荐）
./test/run_tests.sh

# 手动运行所有单元测试
cd backend && go test ./... -v

# 运行特定层级测试
go test ./internal/model/... -v          # Model 层
go test ./internal/handler/... -v        # Handler 层
go test ./internal/service/... -v        # Service 层
go test ./internal/repository/... -v     # Repository 层
go test ./internal/middleware/... -v      # Middleware 层
go test ./internal/utils/... -v          # Utils 层

# 查看覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

#### 集成测试

```bash
# 运行所有集成测试（需要本地 MySQL）
go test -tags integration -v ./internal/repository/integration/...

# 运行特定集成测试
go test -tags integration -v ./internal/repository/integration/... -run TestUserRepository_Integration
go test -tags integration -v ./internal/repository/integration/... -run TestTransactionRepository_Integration
go test -tags integration -v ./internal/repository/integration/... -run TestPortfolioRepository_Integration

# 使用测试脚本运行（包含集成测试）
./test/run_tests.sh --with-integration
```

**集成测试前置条件：**
- 本地 MySQL 服务运行
- 测试数据库 `stock_analysis_test` 已创建
- 数据库连接信息配置正确（默认：root/soyorin114@localhost:3306）

---

## 7. 测试环境

| 项目 | 配置 |
|------|------|
| 操作系统 | macOS (darwin/arm64) |
| Go 版本 | 1.26.1 |
| 测试框架 | Go testing |
| HTTP 测试 | httptest |
| Mock 方式 | 内存存储模拟 + 接口实现 |
| 数据库驱动 | gorm.io/driver/sqlite (单元测试) / gorm.io/driver/mysql (集成测试) |
| 集成测试数据库 | MySQL 8.0 (stock_analysis_test) |

---

**报告生成时间**: 2026-05-15 15:30
