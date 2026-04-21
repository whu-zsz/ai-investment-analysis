# 后端单元测试报告

> **项目**: AI 投资分析系统后端
> **执行日期**: 2026-04-20
> **执行人**: Claude AI
> **Go 版本**: 1.26.1
> **测试状态**: ✅ **全部通过**

---

## 1. 测试概况

| 模块 | 测试文件 | 用例数 | 通过 | 失败 | 覆盖率 |
|------|----------|--------|------|------|--------|
| utils/crypto | `internal/utils/crypto_test.go` | 8 | 8 | 0 | 100% |
| utils/jwt | `internal/utils/jwt_test.go` | 10 | 10 | 0 | 92.9% |
| middleware/auth | `internal/middleware/auth_test.go` | 6 | 6 | 0 | 69.0% |
| handler/user | `internal/handler/user_test.go` | 8 | 8 | 0 | 49.6% |
| handler/transaction | `internal/handler/transaction_test.go` | 13 | 13 | 0 | 49.6% |
| handler/upload | `internal/handler/upload_test.go` | 8 | 8 | 0 | 49.6% |
| handler/portfolio | `internal/handler/portfolio_test.go` | 5 | 5 | 0 | 49.6% |
| service/user | `internal/service/user_service_test.go` | 11 | 11 | 0 | 14.2% |
| service/transaction | `internal/service/transaction_service_test.go` | 18 | 18 | 0 | 14.2% |
| **总计** | **9 个文件** | **87** | **87** | **0** | **24.1%** |

### 测试状态

✅ **全部通过** - 87 个测试用例，0 个失败

---

## 2. 测试详情

### 2.1 utils/crypto_test.go (密码工具测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestHashPassword_Success | 密码哈希生成成功 | ✅ PASS |
| TestHashPassword_DifferentPasswords | 不同密码生成不同哈希 | ✅ PASS |
| TestHashPassword_SamePasswordDifferentHash | 相同密码生成不同哈希（bcrypt盐） | ✅ PASS |
| TestCheckPassword_CorrectPassword | 正确密码验证 | ✅ PASS |
| TestCheckPassword_WrongPassword | 错误密码验证 | ✅ PASS |
| TestCheckPassword_EmptyPassword | 空密码验证 | ✅ PASS |
| TestCheckPassword_InvalidHash | 无效哈希验证 | ✅ PASS |
| TestHashPassword_EmptyPassword | 空密码哈希 | ✅ PASS |

**覆盖率**: 100%

### 2.2 utils/jwt_test.go (JWT 工具测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestGenerateToken_Success | Token 生成成功 | ✅ PASS |
| TestParseToken_Success | Token 解析成功 | ✅ PASS |
| TestParseToken_InvalidToken | 无效 Token 解析 | ✅ PASS |
| TestParseToken_EmptyToken | 空 Token 解析 | ✅ PASS |
| TestParseToken_WrongSecret | 错误密钥解析 | ✅ PASS |
| TestParseToken_TamperedToken | 篡改 Token 解析 | ✅ PASS |
| TestGenerateToken_DifferentSecrets | 不同密钥生成不同签名 | ✅ PASS |
| TestGenerateToken_DifferentUsers | 不同用户生成不同 Token | ✅ PASS |
| TestGenerateToken_Expiration | Token 过期处理 | ✅ PASS |
| TestTokenRoundTrip | 完整流程测试（4个子测试） | ✅ PASS |

**覆盖率**: 92.9%

### 2.3 middleware/auth_test.go (认证中间件测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestAuthMiddleware_MissingHeader | 缺少 Authorization Header | ✅ PASS |
| TestAuthMiddleware_InvalidFormat | 无效 Authorization 格式（4个子测试） | ✅ PASS |
| TestAuthMiddleware_InvalidToken | 无效 Token | ✅ PASS |
| TestAuthMiddleware_ValidToken | 有效 Token | ✅ PASS |
| TestAuthMiddleware_WrongSecret | 错误密钥签名 | ✅ PASS |
| TestAuthMiddleware_ContextValues | Context 用户信息（3个子测试） | ✅ PASS |

**覆盖率**: 69.0%

### 2.4 handler/user_test.go (用户处理器测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestRegister_Success | 用户注册成功 | ✅ PASS |
| TestRegister_InvalidRequest | 无效注册请求（4个子测试） | ✅ PASS |
| TestRegister_UsernameExists | 用户名已存在 | ✅ PASS |
| TestLogin_Success | 用户登录成功 | ✅ PASS |
| TestLogin_InvalidCredentials | 登录凭证错误 | ✅ PASS |
| TestGetProfile_Success | 获取用户信息成功 | ✅ PASS |
| TestGetProfile_UserNotFound | 用户不存在 | ✅ PASS |
| TestUpdateProfile_Success | 更新用户信息成功 | ✅ PASS |
| TestLogout_Success | 登出成功 | ✅ PASS |

**覆盖率**: 49.6%

### 2.5 handler/transaction_test.go (交易处理器测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestCreateTransaction_Success | 创建交易成功 | ✅ PASS |
| TestCreateTransaction_InvalidRequest | 无效创建请求（3个子测试） | ✅ PASS |
| TestCreateTransaction_ServiceError | 服务层错误 | ✅ PASS |
| TestGetTransactions_Success | 获取交易列表成功 | ✅ PASS |
| TestGetTransactions_DefaultPagination | 默认分页 | ✅ PASS |
| TestGetTransaction_Success | 获取交易详情成功 | ✅ PASS |
| TestGetTransaction_NotFound | 交易不存在 | ✅ PASS |
| TestGetTransaction_InvalidID | 无效 ID | ✅ PASS |
| TestUpdateTransaction_Success | 更新交易成功 | ✅ PASS |
| TestUpdateTransaction_NotFound | 更新不存在的交易 | ✅ PASS |
| TestDeleteTransaction_Success | 删除交易成功 | ✅ PASS |
| TestDeleteTransaction_NotFound | 删除不存在的交易 | ✅ PASS |
| TestGetTransactionStats_Success | 获取交易统计成功 | ✅ PASS |
| TestGetTransactionStats_Error | 获取交易统计错误 | ✅ PASS |

**覆盖率**: 49.6%

### 2.6 handler/upload_test.go (上传处理器测试) ⭐ 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestUploadFile_Success | 上传文件成功 | ✅ PASS |
| TestUploadFile_NoFile | 没有上传文件 | ✅ PASS |
| TestUploadFile_UnsupportedType | 不支持的文件类型 | ✅ PASS |
| TestUploadFile_ExcelFile | 上传 Excel 文件 | ✅ PASS |
| TestUploadFile_ServiceError | 服务层错误 | ✅ PASS |
| TestGetUploadHistory_Success | 获取上传历史成功 | ✅ PASS |
| TestGetUploadHistory_Empty | 空上传历史 | ✅ PASS |
| TestGetUploadHistory_ServiceError | 服务层错误 | ✅ PASS |

**覆盖率**: 49.6%

### 2.7 handler/portfolio_test.go (持仓处理器测试) ⭐ 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestGetPortfolios_Success | 获取持仓列表成功 | ✅ PASS |
| TestGetPortfolios_Empty | 空持仓列表 | ✅ PASS |
| TestGetPortfolios_MultipleAssets | 多个持仓 | ✅ PASS |
| TestGetPortfolios_ServiceError | 服务层错误 | ✅ PASS |
| TestGetPortfolios_WithProfit | 带盈亏的持仓 | ✅ PASS |

**覆盖率**: 49.6%

### 2.8 service/user_service_test.go (用户服务测试)

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestUserService_Register_Success | 用户注册成功 | ✅ PASS |
| TestUserService_Register_UsernameExists | 用户名已存在 | ✅ PASS |
| TestUserService_Register_EmailExists | 邮箱已存在 | ✅ PASS |
| TestUserService_Login_Success | 用户登录成功 | ✅ PASS |
| TestUserService_Login_WrongPassword | 密码错误 | ✅ PASS |
| TestUserService_Login_UserNotFound | 用户不存在 | ✅ PASS |
| TestUserService_GetProfile_Success | 获取用户信息成功 | ✅ PASS |
| TestUserService_GetProfile_UserNotFound | 用户不存在 | ✅ PASS |
| TestUserService_UpdateProfile_Success | 更新用户信息成功 | ✅ PASS |
| TestUserService_UpdateProfile_UserNotFound | 更新不存在的用户 | ✅ PASS |
| TestUserService_Login_InactiveUser | 用户账户已停用 | ✅ PASS |
| TestUserService_Register_DatabaseError | 数据库错误 | ✅ PASS |

**覆盖率**: 14.2%

### 2.9 service/transaction_service_test.go (交易服务测试) ⭐ 新增

| 测试用例 | 描述 | 结果 |
|----------|------|------|
| TestTransactionService_CreateTransaction_Success | 创建交易成功 | ✅ PASS |
| TestTransactionService_CreateTransaction_InvalidDate | 无效日期格式 | ✅ PASS |
| TestTransactionService_CreateTransaction_InvalidQuantity | 无效数量 | ✅ PASS |
| TestTransactionService_CreateTransaction_InvalidPrice | 无效价格 | ✅ PASS |
| TestTransactionService_GetTransactions_Success | 获取交易列表成功 | ✅ PASS |
| TestTransactionService_GetTransactions_Pagination | 分页测试 | ✅ PASS |
| TestTransactionService_GetTransactions_DefaultPage | 默认分页 | ✅ PASS |
| TestTransactionService_GetTransactionByID_Success | 获取交易详情成功 | ✅ PASS |
| TestTransactionService_GetTransactionByID_NotFound | 交易不存在 | ✅ PASS |
| TestTransactionService_GetTransactionByID_WrongUser | 其他用户的交易 | ✅ PASS |
| TestTransactionService_UpdateTransaction_Success | 更新交易成功 | ✅ PASS |
| TestTransactionService_UpdateTransaction_NotFound | 更新不存在的交易 | ✅ PASS |
| TestTransactionService_DeleteTransaction_Success | 删除交易成功 | ✅ PASS |
| TestTransactionService_DeleteTransaction_NotFound | 删除不存在的交易 | ✅ PASS |
| TestTransactionService_GetTransactionStats_Success | 获取交易统计成功 | ✅ PASS |
| TestTransactionService_CreateTransaction_RepositoryError | 仓储层错误 | ✅ PASS |

**覆盖率**: 14.2%

---

## 3. 覆盖率详情

### 3.1 已覆盖函数

| 文件 | 函数 | 覆盖率 |
|------|------|--------|
| internal/utils/crypto.go | HashPassword | 100% |
| internal/utils/crypto.go | CheckPassword | 100% |
| internal/utils/jwt.go | GenerateToken | 100% |
| internal/utils/jwt.go | ParseToken | 85.7% |
| internal/middleware/auth.go | AuthMiddleware | 100% |
| internal/handler/user.go | NewUserHandler | 100% |
| internal/handler/user.go | Register | 100% |
| internal/handler/user.go | Login | 77.8% |
| internal/handler/user.go | Logout | 100% |
| internal/handler/user.go | GetProfile | 100% |
| internal/handler/user.go | UpdateProfile | 60.0% |
| internal/handler/transaction.go | CreateTransaction | 100% |
| internal/handler/transaction.go | GetTransactions | 100% |
| internal/handler/transaction.go | GetTransaction | 100% |
| internal/handler/transaction.go | UpdateTransaction | 100% |
| internal/handler/transaction.go | DeleteTransaction | 100% |
| internal/handler/transaction.go | GetTransactionStats | 100% |
| internal/handler/upload.go | UploadFile | 100% |
| internal/handler/upload.go | GetUploadHistory | 100% |
| internal/handler/portfolio.go | GetPortfolios | 100% |
| internal/service/user_service.go | Register | 90.9% |
| internal/service/user_service.go | Login | 91.7% |
| internal/service/user_service.go | GetProfile | 100% |
| internal/service/user_service.go | UpdateProfile | 83.3% |

### 3.2 未覆盖模块

以下模块尚未编写测试：

- `internal/repository/` - 数据访问层
- `internal/service/ai_service.go` - AI 分析服务
- `internal/service/upload_service.go` - 上传服务
- `internal/service/market_*.go` - 市场数据服务
- `internal/handler/analysis.go` - AI 分析处理器
- `internal/handler/market.go` - 市场数据处理器

---

## 4. 运行命令

### 4.1 基本测试命令

```bash
# 进入后端目录
cd /Users/lnm/Downloads/stock_whu/ai-investment-analysis/backend

# 运行所有单元测试
go test ./internal/utils/... ./internal/middleware/... ./internal/handler/... ./internal/service/... -v

# 运行特定模块测试
go test ./internal/utils/... -v
go test ./internal/middleware/... -v
go test ./internal/handler/... -v
go test ./internal/service/... -v
```

### 4.2 查看测试覆盖率

```bash
# 查看各模块覆盖率
go test ./internal/utils/... ./internal/middleware/... ./internal/handler/... ./internal/service/... -cover

# 输出示例:
# ok      stock-analysis-backend/internal/utils       coverage: 92.9% of statements
# ok      stock-analysis-backend/internal/middleware  coverage: 69.0% of statements
# ok      stock-analysis-backend/internal/handler     coverage: 49.6% of statements
# ok      stock-analysis-backend/internal/service     coverage: 14.2% of statements
```

### 4.3 生成覆盖率报告

```bash
# 生成覆盖率文件
go test ./internal/utils/... ./internal/middleware/... ./internal/handler/... ./internal/service/... -coverprofile=coverage.out

# 查看函数级别覆盖率
go tool cover -func=coverage.out

# 输出示例:
# stock-analysis-backend/internal/utils/crypto.go:7:    HashPassword      100.0%
# stock-analysis-backend/internal/utils/crypto.go:12:   CheckPassword     100.0%
# stock-analysis-backend/internal/utils/jwt.go:16:      GenerateToken     100.0%
# ...
# total:                                                 (statements)      24.1%

# 生成 HTML 覆盖率报告（可在浏览器中查看）
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS 打开 HTML 文件
```

### 4.4 运行单个测试用例

```bash
# 运行特定测试函数
go test ./internal/utils/... -run TestHashPassword_Success -v

# 运行匹配模式的测试
go test ./internal/service/... -run "TestUserService_Login" -v

# 输出示例:
# === RUN   TestUserService_Login_Success
# --- PASS: TestUserService_Login_Success (0.21s)
# === RUN   TestUserService_Login_WrongPassword
# --- PASS: TestUserService_Login_WrongPassword (0.21s)
# PASS
```

### 4.5 测试输出说明

```bash
# -v 参数显示详细输出
go test ./internal/service/... -v

# 输出示例:
# === RUN   TestUserService_Register_Success
# --- PASS: TestUserService_Register_Success (0.10s)
# === RUN   TestUserService_Register_UsernameExists
# --- PASS: TestUserService_Register_UsernameExists (0.00s)
# PASS
# coverage: 14.2% of statements
# ok      stock-analysis-backend/internal/service     1.398s

# 输出含义:
# === RUN    - 测试开始运行
# --- PASS   - 测试通过
# --- FAIL   - 测试失败（会显示错误信息）
# (0.10s)   - 测试耗时
# coverage  - 覆盖率百分比
# ok        - 包测试通过
# FAIL      - 包测试失败
```

### 4.6 完整测试执行流程

```bash
# 1. 进入项目目录
cd /Users/lnm/Downloads/stock_whu/ai-investment-analysis/backend

# 2. 运行全部测试（带详细输出和覆盖率）
go test ./internal/utils/... ./internal/middleware/... ./internal/handler/... ./internal/service/... -v -cover

# 3. 生成覆盖率报告
go test ./internal/utils/... ./internal/middleware/... ./internal/handler/... ./internal/service/... -coverprofile=coverage.out

# 4. 查看总覆盖率
go tool cover -func=coverage.out | grep "total:"

# 输出:
# total:                                      (statements)            24.1%

# 5. 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

### 4.7 常见问题排查

```bash
# 如果测试失败，查看详细错误
go test ./internal/service/... -v -run "失败的测试名"

# 清理测试缓存
go clean -testcache

# 重新运行测试
go test ./... -count=1
```

---

## 5. 测试文件位置

```
backend/
├── internal/
│   ├── utils/
│   │   ├── crypto.go
│   │   ├── crypto_test.go         ← 密码工具测试
│   │   ├── jwt.go
│   │   └── jwt_test.go            ← JWT 工具测试
│   ├── middleware/
│   │   ├── auth.go
│   │   └── auth_test.go           ← 认证中间件测试
│   ├── handler/
│   │   ├── user.go
│   │   ├── user_test.go           ← 用户处理器测试
│   │   ├── transaction.go
│   │   ├── transaction_test.go    ← 交易处理器测试
│   │   ├── upload.go
│   │   ├── upload_test.go         ← 上传处理器测试 ⭐ 新增
│   │   ├── portfolio.go
│   │   └── portfolio_test.go      ← 持仓处理器测试 ⭐ 新增
│   └── service/
│       ├── user_service.go
│       ├── user_service_test.go   ← 用户服务测试
│       ├── transaction_service.go
│       └── transaction_service_test.go ← 交易服务测试 ⭐ 新增
```

---

## 6. 后续建议

### 6.1 优先级高 ✅ 已完成

1. ✅ **Service 层测试** - `user_service_test.go`, `transaction_service_test.go`
2. ✅ **Transaction Handler 测试** - `transaction_test.go`
3. **Repository 层测试** - 需要集成测试数据库或使用 mock

### 6.2 优先级中 ✅ 已完成

4. ✅ **Transaction Service 测试** - `transaction_service_test.go`
5. ✅ **Upload Handler 测试** - `upload_test.go`
6. ✅ **Portfolio Handler 测试** - `portfolio_test.go`
7. **Market Handler 测试** - 测试市场数据获取

### 6.3 优先级低

8. **AI Analysis Handler 测试** - 需要 mock LLM API
9. **集成测试** - 使用真实数据库进行端到端测试
10. **性能测试** - API 并发压力测试

---

## 7. 测试环境

| 项目 | 配置 |
|------|------|
| 操作系统 | macOS (darwin/arm64) |
| Go 版本 | 1.26.1 |
| 测试框架 | Go testing |
| HTTP 测试 | httptest |
| Mock 方式 | 手动 Mock 接口实现 |

---

**报告生成时间**: 2026-04-20 22:10
