# Stock Analysis Backend

基于Go语言的投资记录分析与预测后端API服务

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Code Size](https://img.shields.io/github/languages/code-size/yourusername/stock-analysis-backend.svg)](https://github.com/yourusername/stock-analysis-backend)

## 📋 项目简介

这是一个基于AI大模型（Deepseek）的投资记录分析与预测系统后端，提供完整的投资记录管理、持仓计算、AI智能分析等功能。

## 📚 快速链接

- **[项目概览](PROJECT_OVERVIEW.md)** - 完整的项目文档
- **[Git设置指南](GIT_SETUP.md)** - Git使用说明
- **[API文档](http://localhost:8080/swagger/index.html)** - Swagger文档（运行后访问）
- **[需求分析](../需求分析与数据库设计.md)** - 原始需求文档

### 核心功能

- ✅ 用户认证与授权（JWT）
- ✅ 文件上传与解析（CSV/Excel）
- ✅ 交易记录管理
- ✅ 持仓自动计算
- ✅ AI投资分析（Deepseek集成）
- ✅ RESTful API
- ✅ Swagger API文档

## 🛠️ 技术栈

- **框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **认证**: JWT
- **配置管理**: Viper
- **日志**: Zap
- **API文档**: Swagger
- **AI集成**: Deepseek API
- **Excel处理**: excelize
- **CSV处理**: gocsv

## 📦 项目结构

```
stock-analysis-backend/
├── cmd/server/main.go          # 入口文件
├── internal/
│   ├── config/                 # 配置管理
│   ├── middleware/             # 中间件（认证、CORS）
│   ├── handler/                # HTTP处理器
│   ├── service/                # 业务逻辑层
│   ├── repository/             # 数据访问层
│   ├── model/                  # 数据模型
│   ├── dto/                    # 数据传输对象
│   ├── utils/                  # 工具函数
│   └── router/                 # 路由配置
├── pkg/                        # 公共包
│   ├── logger/                 # 日志
│   ├── response/               # 统一响应
│   └── deepseek/               # AI客户端
├── uploads/                    # 上传文件目录
├── docs/                       # Swagger文档
├── .env                        # 环境变量
├── .env.example                # 环境变量示例
└── README.md
```

## 🚀 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+
- Deepseek API Key

### 安装步骤

1. **克隆项目**
```bash
cd /mnt/windata/stock/stock-analysis-backend
```

2. **安装依赖**
```bash
go mod download
```

3. **配置环境变量**
```bash
cp .env.example .env
# 编辑.env文件，填入实际配置
```

4. **创建数据库**
```sql
CREATE DATABASE stock_analysis CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

5. **运行服务**
```bash
go run cmd/server/main.go
```

6. **访问Swagger文档**
```
http://localhost:8080/swagger/index.html
```

## 📝 环境变量配置

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| DB_HOST | 数据库地址 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户名 | root |
| DB_PASSWORD | 数据库密码 | your_password |
| DB_NAME | 数据库名 | stock_analysis |
| JWT_SECRET | JWT密钥 | your_secret_key |
| JWT_EXPIRE_HOURS | Token有效期（小时） | 24 |
| DEEPSEEK_API_KEY | Deepseek API密钥 | your_api_key |
| DEEPSEEK_API_URL | Deepseek API地址 | https://api.deepseek.com |
| UPLOAD_PATH | 文件上传路径 | ./uploads |
| MAX_UPLOAD_SIZE | 最大文件大小（字节） | 10485760 |

## 📚 API接口文档

### 认证接口

#### 用户注册
```
POST /api/v1/auth/register
```

#### 用户登录
```
POST /api/v1/auth/login
```

### 用户接口

#### 获取用户信息
```
GET /api/v1/user/profile
Authorization: Bearer <token>
```

#### 更新用户信息
```
PUT /api/v1/user/profile
Authorization: Bearer <token>
```

### 文件上传接口

#### 上传投资记录
```
POST /api/v1/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

#### 获取上传历史
```
GET /api/v1/upload/history
Authorization: Bearer <token>
```

### 交易记录接口

#### 创建交易记录
```
POST /api/v1/transactions
Authorization: Bearer <token>
```

#### 获取交易记录列表
```
GET /api/v1/transactions?page=1&page_size=20
Authorization: Bearer <token>
```

#### 获取交易统计
```
GET /api/v1/transactions/stats
Authorization: Bearer <token>
```

#### 删除交易记录
```
DELETE /api/v1/transactions/:id
Authorization: Bearer <token>
```

### 持仓接口

#### 获取持仓列表
```
GET /api/v1/portfolios
Authorization: Bearer <token>
```

### AI分析接口

#### 生成投资总结
```
POST /api/v1/analysis/summary?start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer <token>
```

#### 获取历史报告
```
GET /api/v1/analysis/reports?report_type=summary&limit=10
Authorization: Bearer <token>
```

## 📊 数据库表结构

### users（用户表）
- 用户基本信息
- 投资偏好
- 累计盈亏

### transactions（交易记录表）
- 交易日期、类型
- 资产信息
- 数量、价格、金额
- 盈亏记录

### portfolios（持仓明细表）
- 持仓数量
- 平均成本
- 当前市值
- 浮动盈亏

### ai_analysis_reports（AI分析报告表）
- 报告类型
- 分析结果
- AI建议
- 图表数据

### uploaded_files（上传文件记录表）
- 文件信息
- 上传状态
- 导入记录数

## 📄 文件格式要求

### CSV格式
```csv
transaction_date,transaction_type,asset_type,asset_code,quantity,price_per_unit,commission
2024-01-15,buy,stock,600519,10,1800.00,18.00
2024-02-20,sell,stock,600519,5,1900.00,9.50
```

### Excel格式
同样结构，第一行为标题行

## 🔒 安全性

- 密码使用bcrypt加密
- JWT Token认证
- SQL注入防护（GORM参数化查询）
- CORS跨域配置
- 文件上传类型和大小限制

## 🧪 测试

### 运行测试脚本
```bash
bash scripts/test_api.sh
```

### 使用Postman
导入 `docs/postman_collection.json`

## 📈 性能优化建议

1. **数据库索引优化**
   - 已添加关键字段索引
   - 建议对大表进行分区

2. **缓存策略**
   - 使用Redis缓存AI分析结果
   - 缓存用户持仓信息

3. **并发处理**
   - 文件解析使用goroutine
   - 批量插入优化

## 🐛 常见问题

### 1. 数据库连接失败
- 检查MySQL是否启动
- 验证数据库配置是否正确
- 确认数据库已创建

### 2. JWT Token无效
- 检查JWT_SECRET配置
- 确认Token未过期
- 验证Authorization Header格式

### 3. 文件上传失败
- 检查uploads目录权限
- 验证文件大小限制
- 确认文件格式正确

## 📞 技术支持

如有问题，请提交Issue或联系开发团队。

## 📜 许可证

Apache 2.0 License

---

**开发团队**: 张盛哲、顾晨旻、林润民
**课程**: 计算机综合项目实践
**指导教师**: 谭小琼
**学校**: 武汉大学计算机学院
**日期**: 2026年3月
