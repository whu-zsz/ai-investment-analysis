# Stock Analysis Backend - 项目概览

## 📊 项目统计

- **总文件数**: 51个
- **代码行数**: 3800+行
- **开发时间**: 2小时
- **Git提交**: 2次

## 🎯 项目架构

### 技术栈
```
前端（未开发）
    ↓
后端 API (Go + Gin)
    ↓
数据层 (GORM + MySQL)
    ↓
AI服务 (Deepseek API)
```

### 目录结构
```
stock-analysis-backend/
├── cmd/server/main.go              # 主程序入口
├── internal/                       # 内部代码
│   ├── config/                     # 配置管理 (2个文件)
│   ├── middleware/                 # 中间件 (2个文件)
│   ├── handler/                    # HTTP处理器 (5个文件)
│   ├── service/                    # 业务逻辑 (7个文件)
│   ├── repository/                 # 数据访问 (5个文件)
│   ├── model/                      # 数据模型 (5个文件)
│   ├── dto/                        # DTO对象 (5个文件)
│   ├── utils/                      # 工具函数 (2个文件)
│   └── router/                     # 路由配置 (1个文件)
├── pkg/                            # 公共包
│   ├── logger/                     # 日志系统
│   ├── response/                   # 统一响应
│   └── deepseek/                   # AI客户端
├── scripts/                        # 脚本文件
│   ├── init_db.sql                 # 数据库初始化
│   └── test_api.sh                 # API测试
├── docs/                           # Swagger文档
├── uploads/                        # 上传文件
├── .env.example                    # 环境变量模板
├── Dockerfile                      # Docker构建
├── docker-compose.yml              # 容器编排
├── Makefile                        # 构建脚本
└── README.md                       # 项目文档
```

## 📚 API接口清单

### 1. 认证模块
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

### 2. 用户模块
- `GET /api/v1/user/profile` - 获取用户信息
- `PUT /api/v1/user/profile` - 更新用户信息

### 3. 文件上传模块
- `POST /api/v1/upload` - 上传投资记录文件
- `GET /api/v1/upload/history` - 获取上传历史

### 4. 交易记录模块
- `POST /api/v1/transactions` - 创建交易记录
- `GET /api/v1/transactions` - 获取交易记录列表
- `GET /api/v1/transactions/stats` - 获取交易统计
- `DELETE /api/v1/transactions/:id` - 删除交易记录

### 5. 持仓管理模块
- `GET /api/v1/portfolios` - 获取持仓列表

### 6. AI分析模块
- `POST /api/v1/analysis/summary` - 生成投资总结
- `GET /api/v1/analysis/reports` - 获取历史报告

## 🗄️ 数据库设计

### 表结构
1. **users** - 用户表 (13个字段)
2. **transactions** - 交易记录表 (16个字段)
3. **portfolios** - 持仓明细表 (13个字段)
4. **ai_analysis_reports** - AI分析报告表 (17个字段)
5. **uploaded_files** - 上传文件记录表 (11个字段)

### 关系
- users (1) → (N) transactions
- users (1) → (N) portfolios
- users (1) → (N) ai_analysis_reports
- users (1) → (N) uploaded_files

## 🚀 快速开始

### 方式1：本地运行
```bash
# 1. 安装Go 1.21+
sudo pacman -S go

# 2. 配置环境变量
cp .env.example .env
# 编辑.env文件

# 3. 初始化数据库
mysql -u root -p < scripts/init_db.sql

# 4. 安装依赖
make deps

# 5. 生成Swagger文档
make swagger

# 6. 运行服务
make run
```

### 方式2：Docker运行
```bash
# 1. 启动所有服务
docker-compose up -d

# 2. 查看日志
docker-compose logs -f backend

# 3. 停止服务
docker-compose down
```

## 🧪 测试

### API测试
```bash
# 运行测试脚本
bash scripts/test_api.sh

# 或手动测试
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'
```

### 上传测试
```bash
# 使用测试数据文件
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test_data.csv"
```

## 📦 依赖库

### 核心依赖
```go
github.com/gin-gonic/gin          // Web框架
gorm.io/gorm                      // ORM
gorm.io/driver/mysql              // MySQL驱动
github.com/golang-jwt/jwt/v5      // JWT认证
github.com/spf13/viper            // 配置管理
go.uber.org/zap                   // 日志
github.com/swaggo/gin-swagger     // Swagger
github.com/go-playground/validator/v10  // 验证
golang.org/x/crypto               // 加密
github.com/xuri/excelize/v2       // Excel处理
github.com/gocarina/gocsv         // CSV处理
github.com/shopspring/decimal     // 精确计算
github.com/google/uuid            // UUID生成
```

## 🔒 安全特性

- ✅ 密码bcrypt加密
- ✅ JWT Token认证
- ✅ SQL注入防护（GORM参数化）
- ✅ CORS跨域配置
- ✅ 文件上传限制
- ✅ 环境变量隔离

## 📈 性能优化

### 已实现
- 数据库索引优化
- 连接池配置
- 批量插入优化
- 计算列优化

### 建议优化
- [ ] Redis缓存层
- [ ] API限流
- [ ] 分库分表
- [ ] 异步任务队列
- [ ] CDN加速

## 🐛 已知问题

1. **缺少测试用例** - 需要添加单元测试和集成测试
2. **错误处理不够完善** - 需要更详细的错误信息
3. **日志记录不完整** - 需要添加关键操作日志
4. **缺少监控告警** - 需要接入Prometheus/Grafana

## 🔮 未来计划

### 短期目标（1周）
- [ ] 添加单元测试
- [ ] 完善错误处理
- [ ] 优化日志系统
- [ ] 添加CI/CD配置

### 中期目标（1个月）
- [ ] 接入Redis缓存
- [ ] 实现实时数据推送
- [ ] 添加更多AI分析功能
- [ ] 优化数据库查询性能

### 长期目标（3个月）
- [ ] 微服务化改造
- [ ] 分布式部署
- [ ] 高可用架构
- [ ] 完整的监控体系

## 📝 开发日志

### 2026-03-16
- ✅ 项目初始化
- ✅ 完成核心代码框架
- ✅ 数据库设计完成
- ✅ Git仓库创建
- ✅ 添加开发工具
- ✅ Docker支持
- ✅ 文档完善

## 👥 贡献者

- **张盛哲** - 后端开发
- **顾晨旻** - 前端开发（待开发）
- **林润民** - 测试与文档

## 📄 许可证

Apache 2.0 License

---

**项目状态**: ✅ 核心功能已完成，可以开始测试

**最后更新**: 2026-03-16
