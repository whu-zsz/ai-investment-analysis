# AI 投资分析系统测试计划

> **项目**: Stock Analysis Backend
> **技术栈**: Go 1.21+ / Gin / GORM / MySQL 8.0 / JWT / DeepSeek & 豆包 AI
> **开发团队**: 张盛哲、顾晨旻、林润民
> **更新日期**: 2026-04-24

---

## 0. 测试代码更新日志

### 2026-04-24 更新记录 (续)

#### 第七批测试代码 (Repository + Service) - 17:30

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 16:30 | `internal/repository/portfolio_repo_test.go` | 持仓仓储测试 | 9 | ✅ 通过 |
| 16:45 | `internal/repository/uploaded_file_repo_test.go` | 上传文件仓储测试 | 7 | ✅ 通过 |
| 17:00 | `internal/service/file_parser_test.go` | 文件解析服务测试 | 9 | ✅ 通过 |
| 17:15 | `internal/repository/analysis_task_repo_test.go` | 分析任务仓储测试 | 13 | ✅ 通过 |
| 17:30 | `internal/repository/analysis_report_repo_test.go` | 分析报告仓储测试 | 12 | ✅ 通过 |

**测试内容：**
- PortfolioRepository: 创建/查找/更新/删除/价格更新
- UploadedFileRepository: 创建/查找/状态更新
- FileParserService: CSV解析/错误处理/金额计算
- AnalysisTaskRepository: 创建/查询/进度更新/运行状态检查
- AnalysisReportRepository: 创建/查询/报告明细/删除

#### 测试覆盖进度

```
第七批: repository + service/file_parser (50 用例)
├── portfolio_repo_test.go        ████████░░  31.9%
├── uploaded_file_repo_test.go    ████████░░  31.9%
├── file_parser_test.go           █████░░░░░  52.7%
├── analysis_task_repo_test.go    ████████░░  31.9%
└── analysis_report_repo_test.go  ████████░░  31.9%

Repository 覆盖率: 0% → 31.9% (新增)
Service 覆盖率:   48.7% → 52.7% (提升 4%)
总计: 21 个测试文件, 249 个用例, 100% 通过率
```

#### Mock 实现说明

**InMemoryPortfolioRepository** (`portfolio_repo_test.go`)
```go
type InMemoryPortfolioRepository struct {
    portfolios map[uint64]*model.Portfolio
    nextID     uint64
}
// 实现: Create, FindByID, FindByUserID, FindByUserAndAsset, Update, Delete, UpdateCurrentPrice
```

**InMemoryUploadedFileRepository** (`uploaded_file_repo_test.go`)
```go
type InMemoryUploadedFileRepository struct {
    files  map[uint64]*model.UploadedFile
    nextID uint64
}
// 实现: Create, FindByID, FindByUserID, UpdateStatus
```

**InMemoryAnalysisTaskRepository** (`analysis_task_repo_test.go`)
```go
type InMemoryAnalysisTaskRepository struct {
    tasks  map[uint64]*model.AnalysisTask
    nextID uint64
}
// 实现: Create, FindByIDAndUserID, FindByUserID, HasRunningTask, UpdateProgress
```

**InMemoryAnalysisReportRepository** (`analysis_report_repo_test.go`)
```go
type InMemoryAnalysisReportRepository struct {
    reports map[uint64]*model.AnalysisReport
    items   map[uint64]*model.AnalysisReportItem
    nextID  uint64
}
// 实现: Create, CreateWithItems, FindByID, FindByIDAndUserID, FindByTaskID, FindByUserID, FindLatestByUser, Delete
```

### 2026-04-24 更新记录 (续)

#### 第六批测试代码 (Service: Upload + Portfolio) - 11:00

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 10:30 | `internal/service/upload_service_test.go` | 上传服务测试 | 11 | ✅ 通过 |
| 11:00 | `internal/service/portfolio_service_test.go` | 持仓服务测试 | 14 | ✅ 通过 |

**测试内容：**
- UploadService: 文件处理/类型验证/大小验证/解析错误/批量创建
- PortfolioService: 持仓获取/买入/卖出/分红/重新计算

#### 测试覆盖进度

```
第六批: service/upload + service/portfolio (25 用例)
├── upload_service_test.go      █████░░░░░  48.7%
└── portfolio_service_test.go   █████░░░░░  48.7%

Service 覆盖率: 38.7% → 48.7% (提升 10%)
总计: 16 个测试文件, 189 个用例, 100% 通过率
```

#### Mock 实现说明

**MockUploadedFileRepository** (`upload_service_test.go`)
```go
type MockUploadedFileRepository struct {
    Files  map[uint64]*model.UploadedFile
    NextID uint64
}
// 实现: Create, FindByID, FindByUserID, UpdateStatus
```

**MockPortfolioRepository** (`portfolio_service_test.go`)
```go
type MockPortfolioRepository struct {
    Portfolios map[uint64]*model.Portfolio
    NextID     uint64
}
// 实现: Create, FindByID, FindByUserID, FindByUserAndAsset, Update, Delete, UpdateCurrentPrice
```

### 2026-04-24 更新记录

#### 第五批测试代码 (Handler: Market + Analysis) - 10:30

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 10:00 | `internal/handler/market_test.go` | 市场数据处理器测试 | 10 | ✅ 通过 |
| 10:30 | `internal/handler/analysis_test.go` | 分析处理器测试 | 18 | ✅ 通过 |

**测试内容：**
- MarketHandler: 快照列表/历史查询/仪表盘数据/时间解析
- AnalysisHandler: 任务创建/任务查询/报告详情/投资总结

#### 测试覆盖进度

```
第五批: handler/market + handler/analysis (28 用例)
├── market_test.go              ████████░░  86.9%
└── analysis_test.go            ████████░░  86.9%

Handler 覆盖率: 49.6% → 86.9% (提升 37.3%)
总计: 14 个测试文件, 162 个用例, 100% 通过率
```

#### Mock 实现说明

**MockMarketSnapshotService** (`market_test.go`)
```go
type MockMarketSnapshotService struct {
    Snapshots        []response.MarketSnapshotResponse
    DashboardSnapshot *response.DashboardMarketSnapshotResponse
    Err              error
}
// 实现: GetLatestSnapshots, GetHistory, GetDashboardSnapshot
```

**MockAIService** (`analysis_test.go`)
```go
type MockAIService struct {
    CreateTaskResult       *response.AnalysisTaskResponse
    GetTaskResult          *response.AnalysisTaskDetailResponse
    GetTasksResult         *response.AnalysisTaskListResponse
    GetReportDetailResult  *response.AnalysisReportDetailResponse
    GenerateSummaryResult  *response.AnalysisReportResponse
    GetReportsResult       []response.AnalysisReportResponse
}
// 实现: CreateStockAnalysisTask, GetAnalysisTask, GetAnalysisTasks,
//       GetAnalysisReportDetail, GenerateInvestmentSummary, GetReports
```

### 2026-04-21 更新记录

#### 第四批测试代码 (Repository + AI Service) - 10:30

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 10:00 | `internal/repository/user_repo_test.go` | 用户仓储测试 | 9 | ✅ 通过 |
| 10:00 | `internal/repository/transaction_repo_test.go` | 交易仓储测试 | 11 | ✅ 通过 |
| 10:30 | `internal/service/ai_service_test.go` | AI 服务测试 | 19 | ✅ 通过 |

**测试内容：**
- UserRepository: 创建/查找/更新/删除/用户名邮箱索引
- TransactionRepository: CRUD/批量创建/分页/资产代码查询/统计
- AIService: 报告列表/任务管理/投资总结生成/日期验证

#### 测试覆盖进度

```
第四批: repository/user + repository/transaction + service/ai (39 用例)
├── user_repo_test.go           (内存 Mock, 9 用例)
├── transaction_repo_test.go    (内存 Mock, 11 用例)
└── ai_service_test.go          █████░░░░░  38.7%

总计: 12 个测试文件, 126 个用例, 100% 通过率, 38.7% service 覆盖率
```

#### Mock 实现说明

**InMemoryUserRepository** (`user_repo_test.go`)
```go
type InMemoryUserRepository struct {
    users    map[uint64]*model.User
    nextID   uint64
    username map[string]*model.User  // 用户名索引
    email    map[string]*model.User  // 邮箱索引
}
```

**InMemoryTransactionRepository** (`transaction_repo_test.go`)
```go
type InMemoryTransactionRepository struct {
    transactions map[uint64]*model.Transaction
    nextID       uint64
}
```

**MockLLMProvider** (`ai_service_test.go`)
```go
type MockLLMProvider struct {
    Content   string
    modelName string
    Err       error
}
// 实现: GetContent, ModelName
```

### 2026-04-20 更新记录

#### 第一批测试代码 (基础测试) - 20:30

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 20:30 | `internal/utils/crypto_test.go` | 密码工具测试 | 8 | ✅ 通过 |
| 20:30 | `internal/utils/jwt_test.go` | JWT 工具测试 | 10 | ✅ 通过 |
| 20:30 | `internal/middleware/auth_test.go` | 认证中间件测试 | 6 | ✅ 通过 |
| 20:30 | `internal/handler/user_test.go` | 用户处理器测试 | 8 | ✅ 通过 |

**测试内容：**
- 密码哈希生成与验证
- JWT Token 生成与解析
- 认证中间件 Header 验证
- 用户注册/登录/资料 CRUD

#### 第二批测试代码 (优先级高) - 21:30

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 21:30 | `internal/service/user_service_test.go` | 用户服务层测试 | 11 | ✅ 通过 |
| 21:30 | `internal/handler/transaction_test.go` | 交易处理器测试 | 13 | ✅ 通过 |

**测试内容：**
- UserService: 注册/登录/资料管理/账户状态
- TransactionHandler: 交易 CRUD/统计/分页

#### 第三批测试代码 (优先级中) - 22:00

| 时间 | 文件 | 描述 | 用例数 | 状态 |
|------|------|------|--------|------|
| 22:00 | `internal/service/transaction_service_test.go` | 交易服务层测试 | 18 | ✅ 通过 |
| 22:00 | `internal/handler/upload_test.go` | 上传处理器测试 | 8 | ✅ 通过 |
| 22:00 | `internal/handler/portfolio_test.go` | 持仓处理器测试 | 5 | ✅ 通过 |

**测试内容：**
- TransactionService: 创建/查询/更新/删除/统计/分页/验证
- UploadHandler: 文件上传/类型验证/历史查询
- PortfolioHandler: 持仓列表/盈亏计算

#### 测试覆盖进度

```
第一批: utils + middleware + handler/user (32 用例)
├── crypto_test.go     ████████████ 100%
├── jwt_test.go        ███████████░  92.9%
├── auth_test.go       ███████░░░░░  69.0%
└── user_test.go       ███░░░░░░░░░  32.9%

第二批: service/user + handler/transaction (24 用例)
├── user_service_test.go   █░░░░░░░░░░   3.8%
└── transaction_test.go    ███░░░░░░░░  32.9%

第三批: service/transaction + handler/upload + handler/portfolio (31 用例)
├── transaction_service_test.go  ██░░░░░░░░░  14.2%
├── upload_test.go               █████░░░░░  49.6%
└── portfolio_test.go            █████░░░░░  49.6%

总计: 9 个测试文件, 87 个用例, 100% 通过率, 24.1% 总覆盖率
```

#### Mock 实现说明

**MockUserRepository** (`user_service_test.go`)
```go
type MockUserRepository struct {
    users         map[uint64]*model.User
    nextID        uint64
    errOnCreate   error
    errOnFindByID error
}
// 实现: Create, FindByID, FindByUsername, FindByEmail, Update, Delete, UpdateLastLogin, UpdateTotalProfit, SetUserActive
```

**MockUserService** (`user_test.go`)
```go
type MockUserService struct {
    RegisterFunc      func(*request.RegisterRequest) (*model.User, error)
    LoginFunc         func(*request.LoginRequest) (*response.LoginResponse, error)
    GetProfileFunc    func(uint64) (*model.User, error)
    UpdateProfileFunc func(uint64, *request.UpdateProfileRequest) (*model.User, error)
}
```

**MockTransactionService** (`transaction_test.go`)
```go
type MockTransactionService struct {
    CreateTransactionFunc  func(userID uint64, req *request.CreateTransactionRequest) error
    GetTransactionsFunc    func(userID uint64, page, pageSize int) (*response.TransactionListResponse, error)
    GetTransactionByIDFunc func(userID uint64, id uint64) (*model.Transaction, error)
    UpdateTransactionFunc  func(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error)
    DeleteTransactionFunc  func(userID uint64, id uint64) error
    GetTransactionStatsFunc func(userID uint64) (*response.TransactionStats, error)
}
```

**MockTransactionRepository** (`transaction_service_test.go`)
```go
type MockTransactionRepository struct {
    transactions    map[uint64]*model.Transaction
    nextID          uint64
    errOnCreate     error
    errOnFindByID   error
    statsResult     *dtoResponse.TransactionStats
}
// 实现: Create, BatchCreate, FindByID, FindByUserID, FindByAssetCode, FindByDateRange, Update, Delete, GetTransactionStats
```

**MockPortfolioService** (`transaction_service_test.go`, `portfolio_test.go`)
```go
type MockPortfolioService struct {
    errOnUpdate bool
}
// 实现: UpdatePortfolioFromTransaction, RecalculatePortfolio, GetPortfolios
```

**MockUploadService** (`upload_test.go`)
```go
type MockUploadService struct {
    ProcessUploadedFileFunc func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error)
    GetUploadHistoryFunc    func(userID uint64) ([]response.UploadHistoryResponse, error)
}
```

---

## 1. 测试背景

本测试计划基于以下实际代码：
- 后端入口: `backend/cmd/server/main.go`
- 后端路由: `backend/internal/router/router.go`
- Handler: `backend/internal/handler/*.go`
- Service: `backend/internal/service/*.go`
- 前端 API: `frontend/src/api/index.ts`

### 1.1 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (React)                      │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP/REST
┌─────────────────────────▼───────────────────────────────────┐
│                   Backend (Go + Gin)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │ Handler  │→ │ Service  │→ │Repository│→ │  MySQL   │    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │
│        │              │                                      │
│        │              ▼                                      │
│        │     ┌──────────────┐                               │
│        │     │ LLM Provider │ ← DeepSeek / 豆包              │
│        │     └──────────────┘                               │
│        │              │                                      │
│        │              ▼                                      │
│        │     ┌──────────────┐                               │
│        │     │Market Data   │ ← Mock / 东方财富              │
│        │     └──────────────┘                               │
└─────────────────────────────────────────────────────────────┘
```

## 2. API 接口矩阵

### 2.1 认证模块 `/api/v1/auth`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /auth/register | POST | username, email, password | 否 | 用户注册 |
| /auth/login | POST | username, password | 否 | 用户登录 |

### 2.2 用户模块 `/api/v1/user`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /user/profile | GET | - | Bearer | 获取用户信息 |
| /user/profile | PUT | phone?, avatar_url?, investment_preference? | Bearer | 更新用户信息 |

### 2.3 上传模块 `/api/v1/upload`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /upload | POST | multipart/form-data (file) | Bearer | 上传投资记录文件 |
| /upload/history | GET | - | Bearer | 获取上传历史 |

### 2.4 交易记录模块 `/api/v1/transactions`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /transactions | POST | transaction_date, transaction_type, asset_code, asset_name, quantity, price_per_unit... | Bearer | 创建交易记录 |
| /transactions | GET | page, page_size, start_date, end_date, asset_code, transaction_type | Bearer | 获取交易列表(分页) |
| /transactions/stats | GET | - | Bearer | 获取交易统计 |
| /transactions/:id | DELETE | - | Bearer | 删除交易记录 |

### 2.5 持仓模块 `/api/v1/portfolios`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /portfolios | GET | - | Bearer | 获取持仓列表 |

### 2.6 Dashboard模块 `/api/v1/dashboard`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /dashboard/market-snapshot | GET | - | Bearer | 获取市场快照 |

### 2.7 市场模块 `/api/v1/market`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /market/snapshots/latest | GET | - | Bearer | 获取最新市场快照 |
| /market/snapshots/history | GET | symbol, limit, start_time, end_time | Bearer | 获取市场快照历史 |

### 2.8 AI分析模块 `/api/v1/analysis`

| 接口 | 方法 | 参数 | 认证 | 功能 |
|------|------|------|------|------|
| /analysis/summary | POST | start_date, end_date (query) | Bearer | 生成投资总结 |
| /analysis/reports | GET | report_type?, limit? | Bearer | 获取历史报告 |

## 3. 测试用例

### 3.1 认证模块测试

#### AUTH_001: 用户注册成功
```
请求: POST /api/v1/auth/register
Body: {"username": "test001", "email": "test001@example.com", "password": "Test123456"}
预期: 200, 返回用户信息
```

#### AUTH_002: 用户注册-用户名重复
```
请求: POST /api/v1/auth/register
Body: {"username": "existing", "email": "test@example.com", "password": "Test123456"}
预期: 400, "username already exists"
```

#### AUTH_003: 用户登录成功
```
请求: POST /api/v1/auth/login
Body: {"username": "test001", "password": "Test123456"}
预期: 200, 返回 token 和 user 信息
```

#### AUTH_004: 用户登录-密码错误
```
请求: POST /api/v1/auth/login
Body: {"username": "test001", "password": "WrongPassword"}
预期: 401, "invalid credentials"
```

#### AUTH_005: 用户登录-用户不存在
```
请求: POST /api/v1/auth/login
Body: {"username": "nonexistent", "password": "Test123456"}
预期: 401, "user not found"
```

### 3.2 用户模块测试

#### USER_001: 获取用户信息
```
请求: GET /api/v1/user/profile
Header: Authorization: Bearer <token>
预期: 200, 返回用户详细信息
```

#### USER_002: 更新用户信息
```
请求: PUT /api/v1/user/profile
Header: Authorization: Bearer <token>
Body: {"phone": "13800138000", "investment_preference": "aggressive"}
预期: 200, 更新成功
```

#### USER_003: 更新用户信息-无效偏好值
```
请求: PUT /api/v1/user/profile
Body: {"investment_preference": "invalid"}
预期: 400, 验证错误
```

### 3.3 上传模块测试

#### UPLOAD_001: 上传CSV文件
```
请求: POST /api/v1/upload
Header: Authorization: Bearer <token>
Body: FormData(file=test.csv)
预期: 200, 返回 file_id, records_imported
```

#### UPLOAD_002: 上传Excel文件
```
请求: POST /api/v1/upload
Header: Authorization: Bearer <token>
Body: FormData(file=test.xlsx)
预期: 200, 返回 file_id, records_imported
```

#### UPLOAD_003: 上传不支持格式
```
请求: POST /api/v1/upload
Header: Authorization: Bearer <token>
Body: FormData(file=test.txt)
预期: 400, "unsupported file type"
```

#### UPLOAD_004: 上传文件过大
```
请求: POST /api/v1/upload
Header: Authorization: Bearer <token>
Body: FormData(file=large_file.csv)
预期: 400, "file size exceeds maximum limit"
```

#### UPLOAD_005: 获取上传历史
```
请求: GET /api/v1/upload/history
Header: Authorization: Bearer <token>
预期: 200, 返回上传记录列表
```

### 3.4 交易记录模块测试

#### TX_001: 创建买入交易
```
请求: POST /api/v1/transactions
Header: Authorization: Bearer <token>
Body: {
  "transaction_date": "2024-01-15",
  "transaction_type": "buy",
  "asset_code": "600519",
  "asset_name": "贵州茅台",
  "quantity": "100",
  "price_per_unit": "1850.00"
}
预期: 200, 创建成功
```

#### TX_002: 创建卖出交易
```
请求: POST /api/v1/transactions
Body: {
  "transaction_date": "2024-02-20",
  "transaction_type": "sell",
  "asset_code": "600519",
  "asset_name": "贵州茅台",
  "quantity": "50",
  "price_per_unit": "1900.00"
}
预期: 200, 创建成功
```

#### TX_003: 创建分红
```
请求: POST /api/v1/transactions
Body: {
  "transaction_date": "2024-03-01",
  "transaction_type": "dividend",
  "asset_code": "600519",
  "asset_name": "贵州茅台",
  "quantity": "0",
  "price_per_unit": "0",
  "total_amount": "5000.00"
}
预期: 200, 创建成功
```

#### TX_004: 创建交易-缺少必填字段
```
请求: POST /api/v1/transactions
Body: {"transaction_type": "buy"}
预期: 400, 验证错误
```

#### TX_005: 创建交易-无效交易类型
```
请求: POST /api/v1/transactions
Body: {
  "transaction_date": "2024-01-15",
  "transaction_type": "invalid",
  "asset_code": "600519",
  "quantity": "100",
  "price_per_unit": "1850.00"
}
预期: 400, 验证错误
```

#### TX_006: 获取交易列表-默认分页
```
请求: GET /api/v1/transactions
Header: Authorization: Bearer <token>
预期: 200, 返回 transactions[], total, page, page_size
```

#### TX_007: 获取交易列表-指定分页
```
请求: GET /api/v1/transactions?page=2&page_size=10
预期: 200, 返回第2页，每页10条
```

#### TX_008: 获取交易列表-日期筛选
```
请求: GET /api/v1/transactions?start_date=2024-01-01&end_date=2024-12-31
���期: 200, 返回指定日期范围内的交易
```

#### TX_009: 获取交易列表-资产筛选
```
请求: GET /api/v1/transactions?asset_code=600519
预期: 200, 返回指定股票的交易
```

#### TX_010: 获取交易列表-类型筛选
```
请求: GET /api/v1/transactions?transaction_type=buy
预期: 200, 返回买入交易
```

#### TX_011: 获取交易统计
```
请求: GET /api/v1/transactions/stats
Header: Authorization: Bearer <token>
预期: 200, 返回 total_transactions, buy_count, sell_count, total_investment, total_profit
```

#### TX_012: 删除交易记录
```
请求: DELETE /api/v1/transactions/1
Header: Authorization: Bearer <token>
预期: 200, 删除成功
```

#### TX_013: 删除不存在的交易
```
请求: DELETE /api/v1/transactions/999999
预期: 404, "transaction not found"
```

### 3.5 持仓模块测试

#### PORT_001: 获取持仓列表
```
请求: GET /api/v1/portfolios
Header: Authorization: Bearer <token>
预期: 200, 返回持仓数组
```

#### PORT_002: 持仓计算验证
```
验证字段: total_quantity, average_cost, market_value, profit_loss, profit_loss_percent
验证逻辑: market_value = total_quantity * current_price
          profit_loss = market_value - (total_quantity * average_cost)
```

### 3.6 Dashboard模块测试

#### DASH_001: 获取市场快照
```
请求: GET /api/v1/dashboard/market-snapshot
Header: Authorization: Bearer <token>
预期: 200, 返回 snapshot_time, indices[], main_chart, stats
```

### 3.7 市场模块测试

#### MKT_001: 获取最新市场快照
```
请求: GET /api/v1/market/snapshots/latest
Header: Authorization: Bearer <token>
预期: 200, 返回市场快照数组
```

#### MKT_002: 获取市场快照历史
```
请求: GET /api/v1/market/snapshots/history?symbol=000001&limit=30
Header: Authorization: Bearer <token>
预期: 200, 返回30条历史记录
```

#### MKT_003: 获取市场快照历史-缺少symbol
```
请求: GET /api/v1/market/snapshots/history
预期: 400, "symbol is required"
```

### 3.8 AI分析模块测试

#### AI_001: 生成投资总结
```
请求: POST /api/v1/analysis/summary?start_date=2024-01-01&end_date=2024-12-31
Header: Authorization: Bearer <token>
预期: 200, 返回分析报告
验证字段: id, report_type, report_title, analysis_period_start, analysis_period_end,
         total_investment, total_profit, profit_rate, risk_level, investment_style,
         summary_text, risk_analysis, pattern_insights, prediction_text, recommendations
```

#### AI_002: 生成投资总结-缺少日期参数
```
请求: POST /api/v1/analysis/summary
预期: 400, "start_date and end_date are required"
```

#### AI_003: 获取历史报告
```
请求: GET /api/v1/analysis/reports
Header: Authorization: Bearer <token>
预期: 200, 返回报告数组
```

#### AI_004: 获取历史报告-指定类型
```
请求: GET /api/v1/analysis/reports?report_type=summary&limit=5
预期: 200, 返回5条summary类型报告
```

## 4. 测试数据准备

### 4.1 测试用户
```sql
-- 用户1
INSERT INTO users (username, email, password_hash, investment_preference, risk_tolerance, total_profit)
VALUES ('testuser1', 'test1@example.com', '$2a$10$...', 'balanced', 'medium', '0.00');

-- 用户2 (已有交易记录)
INSERT INTO users (username, email, password_hash, investment_preference, risk_tolerance, total_profit)
VALUES ('testuser2', 'test2@example.com', '$2a$10$...', 'aggressive', 'high', '50000.00');
```

### 4.2 测试交易数据
```csv
transaction_date,transaction_type,asset_code,asset_name,quantity,price_per_unit,total_amount
2024-01-15,buy,600519,贵州茅台,100,1850.00,185000.00
2024-02-20,sell,600519,贵州茅台,50,1900.00,95000.00
2024-03-10,buy,000858,五粮液,200,180.00,36000.00
2024-04-05,dividend,000858,五粮液,0,0,5000.00
2024-05-01,buy,300750,宁德时���,50,380.00,19000.00
```

### 4.3 测试上传文件格式
```csv
交易日期,交易类型,证券代码,证券名称,成交数量,成交单价,成交金额
2024-01-15,买入,600519,贵州茅台,100,1850.00,185000.00
2024-02-20,卖出,600519,贵州茅台,50,1900.00,95000.00
```

## 5. 测试执行顺序

### 5.1 第一阶段: 认证流程 (AUTH_001 ~ AUTH_005)
1. 注册新用户 -> 登录 -> 使用token

### 5.2 第二阶段: 用户模块 (USER_001 ~ USER_003)
1. 获取个人信息 -> 更新个人信息

### 5.3 第三阶段: 上传模块 (UPLOAD_001 ~ UPLOAD_005)
1. 上传CSV -> 上传Excel -> 获取上传历史

### 5.4 第四阶段: 交易记录 (TX_001 ~ TX_013)
1. 创建交易 -> 查询列表 -> 统计 -> 删除

### 5.5 第五阶段: 持仓 (PORT_001 ~ PORT_002)
1. 获取持仓 -> 验证计算逻辑

### 5.6 第六阶段: Dashboard/Market (DASH_001, MKT_001 ~ MKT_003)
1. Dashboard快照 -> 市场快照

### 5.7 第七阶段: AI分析 (AI_001 ~ AI_004)
1. 生成报告 -> 查询报告

## 6. 验收标准

| 模块 | 通过率 | 关键指标 |
|------|--------|----------|
| 认证 | 100% | 登录/注册/鉴权 |
| 用户 | 100% | 获取/更新 |
| 上传 | 100% | 文件处理 |
| 交易 | 100% | CRUD+分页+筛选 |
| 持仓 | 100% | 收益率计算 |
| Dashboard | 100% | 快照数据 |
| AI分析 | 100% | 报告生成 |

## 7. 测试工具

- 后端: Go testing + testify
- API: curl / Postman / Newman
- 前端: Vitest + React Testing Library

## 8. 后续扩展

- [ ] 性能测试 (高并发API调用)
- [ ] 安全测试 (XSS/SQL注入)
- [ ] 边界条件测试 (空数据/最大值)
- [ ] E2E测试 (Playwright)

---

## 9. 测试环境配置

### 9.1 环境要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 后端运行环境 |
| MySQL | 8.0+ | 数据库 |
| Node.js | 18+ | 前端测试 |

### 9.2 数据库准备

```bash
# 创建测试数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS stock_analysis CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 可选：初始化测试数据
mysql -u root -p stock_analysis < backend/scripts/init_db.sql
```

### 9.3 环境变量配置 (`.env`)

```bash
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=soyorin114
DB_NAME=stock_analysis

# JWT
JWT_SECRET=test_jwt_secret_for_testing_only
JWT_EXPIRE_HOURS=24

# LLM (测试时可使用 mock)
LLM_PROVIDER=deepseek
DEEPSEEK_API_KEY=your_api_key_here
DEEPSEEK_API_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-chat

# Market Data (使用 mock 避免外部依赖)
MARKET_ENABLED=true
MARKET_PROVIDER=mock
MARKET_SYMBOLS=000001.SH,399001.SZ,399006.SZ,000300.SH
MARKET_SNAPSHOT_INTERVAL=60

# Upload
UPLOAD_PATH=./uploads
MAX_UPLOAD_SIZE=10485760
```

---

## 10. 测试执行指南

### 10.1 启动后端服务

```bash
cd backend

# 方式一：直接运行
go run cmd/server/main.go

# 方式二：使用 Makefile
make run

# 方式三：编译后运行
go build -o bin/server cmd/server/main.go
./bin/server
```

### 10.2 运行 API 测试脚本

```bash
# 确保服务已启动 (端口 8080)
bash backend/scripts/test_api.sh
```

### 10.3 使用 curl 手动测试

```bash
# 健康检查
curl http://localhost:8080/health

# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@test.com","password":"Test123456"}'

# 登录 (获取 token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test123456"}'

# 使用 token 访问受保护接口
TOKEN="your_token_here"
curl http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"

# 上传文件
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@backend/test_data.csv"

# 获取交易列表
curl "http://localhost:8080/api/v1/transactions?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 获取持仓
curl http://localhost:8080/api/v1/portfolios \
  -H "Authorization: Bearer $TOKEN"

# AI 分析 (需要配置 API Key)
curl -X POST "http://localhost:8080/api/v1/analysis/summary?start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer $TOKEN"
```

### 10.4 访问 Swagger 文档

服务启动后访问: http://localhost:8080/swagger/index.html

---

## 11. Go 单元测试 (建议新增)

### 11.1 测试文件结构

```
backend/
├── internal/
│   ├── utils/
│   │   ├── crypto.go
│   │   └── crypto_test.go      # 密码加密测试
│   ├── service/
│   │   ├── user.go
│   │   └── user_test.go        # 用户服务测试
│   └── handler/
│       ├── user.go
│       └── user_test.go        # HTTP 处理器测试
└── test/
    ├── integration/             # 集成测试
    └── fixtures/                # 测试数据
```

### 11.2 示例：密码工具测试

```go
// internal/utils/crypto_test.go
package utils

import "testing"

func TestHashPassword(t *testing.T) {
    password := "Test123456"
    hash, err := HashPassword(password)
    if err != nil {
        t.Fatalf("HashPassword failed: %v", err)
    }
    if hash == password {
        t.Error("Hash should not equal plain password")
    }
}

func TestCheckPassword(t *testing.T) {
    password := "Test123456"
    hash, _ := HashPassword(password)
    
    if !CheckPassword(password, hash) {
        t.Error("Correct password should pass")
    }
    if CheckPassword("wrongpassword", hash) {
        t.Error("Wrong password should fail")
    }
}
```

### 11.3 示例：JWT 工具测试

```go
// internal/utils/jwt_test.go
package utils

import "testing"

func TestGenerateAndParseToken(t *testing.T) {
    userID := uint(1)
    secret := "test_secret"
    
    token, err := GenerateToken(userID, secret, 24)
    if err != nil {
        t.Fatalf("GenerateToken failed: %v", err)
    }
    
    parsedID, err := ParseToken(token, secret)
    if err != nil {
        t.Fatalf("ParseToken failed: %v", err)
    }
    if parsedID != userID {
        t.Errorf("Expected %d, got %d", userID, parsedID)
    }
}
```

### 11.4 运行单元测试

```bash
cd backend

# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/utils/...

# 带覆盖率
go test -cover ./...

# 详细输出
go test -v ./...
```

---

## 12. 测试检查清单

### 12.1 冒烟测试 (Smoke Test)

| 序号 | 检查项 | 命令/操作 | 预期结果 |
|------|--------|-----------|----------|
| 1 | 服务启动 | `go run cmd/server/main.go` | 无错误，监听 8080 |
| 2 | 健康检查 | `curl /health` | `{"status": "ok"}` |
| 3 | Swagger | 浏览器访问 `/swagger` | 页面正常显示 |
| 4 | 用户注册 | POST `/auth/register` | 200 OK |
| 5 | 用户登录 | POST `/auth/login` | 返回 token |
| 6 | 认证访问 | GET `/user/profile` + Bearer | 200 OK |

### 12.2 功能测试检查清单

- [ ] **认证模块**
  - [ ] 用户注册成功
  - [ ] 重复用户名注册失败
  - [ ] 登录成功返回 token
  - [ ] 错误密码登录失败
  - [ ] Token 过期处理

- [ ] **用户模块**
  - [ ] 获取用户信息
  - [ ] 更新用户信息
  - [ ] 投资偏好设置

- [ ] **文件上传**
  - [ ] CSV 文件上传解析
  - [ ] Excel 文件上传解析
  - [ ] 不支持格式拒绝
  - [ ] 文件大小限制
  - [ ] 上传历史查询

- [ ] **交易记录**
  - [ ] 创建买入记录
  - [ ] 创建卖出记录
  - [ ] 创建分红记录
  - [ ] 分页查询
  - [ ] 日期筛选
  - [ ] 资产筛选
  - [ ] 交易统计
  - [ ] 删除记录

- [ ] **持仓计算**
  - [ ] 持仓数量正确
  - [ ] 平均成本正确
  - [ ] 盈亏计算正确

- [ ] **市场数据**
  - [ ] Mock 模式正常
  - [ ] 快照数据格式
  - [ ] 历史查询

- [ ] **AI 分析**
  - [ ] 报告生成 (需 API Key)
  - [ ] 历史报告查询

---

## 13. 常见问题排查

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 数据库连接失败 | MySQL 未启动或配置错误 | 检查 MySQL 状态，验证 `.env` 配置 |
| JWT Token 无效 | Secret 不匹配或 Token 过期 | 检查 `JWT_SECRET`，重新登录获取 token |
| 文件上传失败 | 目录权限或文件过大 | 检查 `uploads/` 权限，验证文件大小 |
| AI 分析失败 | API Key 无效 | 检查 `DEEPSEEK_API_KEY` 或 `DOUBAO_API_KEY` |
| CORS 错误 | 跨域配置问题 | 检查 `middleware/cors.go` 配置 |

---

## 14. 测试报告模板

```markdown
# 测试执行报告

**执行日期**: YYYY-MM-DD
**执行人**:
**环境**: 本地开发环境

## 测试概览

| 模块 | 用例数 | 通过 | 失败 | 通过率 |
|------|--------|------|------|--------|
| 认证 | 5 | - | - | -% |
| 用户 | 3 | - | - | -% |
| 上传 | 5 | - | - | -% |
| 交易 | 13 | - | - | -% |
| 持仓 | 2 | - | - | -% |
| 市场 | 3 | - | - | -% |
| AI | 4 | - | - | -% |
| **总计** | **35** | - | - | -% |

## 失败用例详情

| 用例ID | 描述 | 错误信息 | 截图 |
|--------|------|----------|------|
| - | - | - | - |

## 建议

-
```

---

## 15. 单元测试执行报告 (2026-04-20)

### 15.1 测试概况

| 模块 | 测试文件 | 用例数 | 覆盖率 | 状态 |
|------|----------|--------|--------|------|
| utils/crypto | `internal/utils/crypto_test.go` | 8 | 100% | ✅ 通过 |
| utils/jwt | `internal/utils/jwt_test.go` | 10 | 92.9% | ✅ 通过 |
| middleware/auth | `internal/middleware/auth_test.go` | 6 | 69.0% | ✅ 通过 |
| handler/user | `internal/handler/user_test.go` | 8 | 12.3% | ✅ 通过 |
| **总计** | **4 个文件** | **32** | **-** | **✅ 全部通过** |

### 15.2 测试详情

#### utils/crypto_test.go
```
✓ TestHashPassword_Success
✓ TestHashPassword_DifferentPasswords
✓ TestHashPassword_SamePasswordDifferentHash
✓ TestCheckPassword_CorrectPassword
✓ TestCheckPassword_WrongPassword
✓ TestCheckPassword_EmptyPassword
✓ TestCheckPassword_InvalidHash
✓ TestHashPassword_EmptyPassword
```

#### utils/jwt_test.go
```
✓ TestGenerateToken_Success
✓ TestParseToken_Success
✓ TestParseToken_InvalidToken
✓ TestParseToken_EmptyToken
✓ TestParseToken_WrongSecret
✓ TestParseToken_TamperedToken
✓ TestGenerateToken_DifferentSecrets
✓ TestGenerateToken_DifferentUsers
✓ TestGenerateToken_Expiration
✓ TestTokenRoundTrip (含 4 个子测试)
```

#### middleware/auth_test.go
```
✓ TestAuthMiddleware_MissingHeader
✓ TestAuthMiddleware_InvalidFormat (含 4 个子测试)
✓ TestAuthMiddleware_InvalidToken
✓ TestAuthMiddleware_ValidToken
✓ TestAuthMiddleware_WrongSecret
✓ TestAuthMiddleware_ContextValues (含 3 个子测试)
```

#### handler/user_test.go
```
✓ TestRegister_Success
✓ TestRegister_InvalidRequest (含 4 个子测试)
✓ TestRegister_UsernameExists
✓ TestLogin_Success
✓ TestLogin_InvalidCredentials
✓ TestGetProfile_Success
✓ TestGetProfile_UserNotFound
✓ TestUpdateProfile_Success
✓ TestLogout_Success
```

### 15.3 运行命令

```bash
# 运行所有测试
go test ./... -v

# 运行特定包测试
go test ./internal/utils/... -v
go test ./internal/middleware/... -v
go test ./internal/handler/... -v

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out -o coverage.html
```

### 15.4 后续测试建议

1. **Service 层测试** - 需要添加 `user_service_test.go` 等
2. **Repository 层测试** - 需要集成测试数据库
3. **其他 Handler 测试** - transaction、portfolio、analysis 等
4. **集成测试** - 使用真实的数据库连接进行端到端测试