# Stock Analysis Backend

基于 Go 语言的投资记录分析与预测后端 API 服务。

## 项目简介

这是一个基于 AI 大模型（支持 DeepSeek / 豆包）的投资记录分析与预测系统后端，提供投资记录管理、持仓计算、AI 智能分析等能力。

## 技术栈

- 框架: Gin
- ORM: GORM
- 数据库: MySQL 8.0
- 认证: JWT
- 配置管理: Viper
- 日志: Zap
- API 文档: Swagger
- AI 集成: DeepSeek / 豆包 API

## 快速开始

### 推荐的本地开发方式

推荐使用：
- **数据库通过 Docker Compose 启动**
- **后端服务在宿主机直接运行**

这样可以避免本机 MySQL 服务状态不稳定、端口冲突和重复初始化问题。

### 1. 安装依赖

```bash
go mod download
```

### 2. 准备环境变量

```bash
cp .env.example .env
```

然后编辑 `.env`。

> 宿主机运行后端时，`DB_HOST` 建议使用 `127.0.0.1`，不要使用 `localhost`，避免部分系统把它解析到 `::1` 导致 MySQL 连接失败。

### 3. 启动数据库

```bash
docker compose up -d mysql
```

首次启动会自动执行 `scripts/init_db.sql`，只创建：
- `stock_analysis`
- `stock_analysis_test`

应用表结构由后端启动时自动迁移生成。

### 4. 启动后端

你可以在 `backend/` 目录运行：

```bash
go run cmd/server/main.go
```

也可以从仓库根目录运行同样命令；当前配置加载逻辑已兼容这两种场景。

### 5. 访问接口文档

```text
http://localhost:8080/swagger/index.html
```

## 环境变量说明

默认使用 `deepseek` 作为 LLM provider。若切换到豆包，请设置：

```bash
LLM_PROVIDER=doubao
DOUBAO_API_KEY=你的方舟 API Key
DOUBAO_API_URL=https://ark.cn-beijing.volces.com
DOUBAO_MODEL=你的模型 ID 或 Endpoint ID
```

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| DB_HOST | 数据库地址（宿主机运行建议 127.0.0.1） | 127.0.0.1 |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户名 | root |
| DB_PASSWORD | 数据库密码 | your_password |
| DB_NAME | 数据库名 | stock_analysis |
| JWT_SECRET | JWT 密钥 | your_secret_key |
| JWT_EXPIRE_HOURS | Token 有效期（小时） | 24 |
| LLM_PROVIDER | LLM 提供方（deepseek/doubao） | deepseek |
| DEEPSEEK_API_KEY | DeepSeek API 密钥 | your_api_key |
| DEEPSEEK_API_URL | DeepSeek API 地址 | https://api.deepseek.com |
| DEEPSEEK_MODEL | DeepSeek 模型名 | deepseek-chat |
| DOUBAO_API_KEY | 豆包 / 方舟 API 密钥 | your_api_key |
| DOUBAO_API_URL | 豆包 / 方舟 API 地址 | https://ark.cn-beijing.volces.com |
| DOUBAO_MODEL | 豆包模型 ID 或 Endpoint ID | ep-xxxx |
| UPLOAD_PATH | 文件上传路径 | ./uploads |
| MAX_UPLOAD_SIZE | 最大文件大小（字节） | 10485760 |

## 数据库与重启说明

### 本地重复启动后端时的建议

1. 优先保证 MySQL 已启动
2. 宿主机运行后端时使用 `DB_HOST=127.0.0.1`
3. 不要反复手动执行旧版建表 SQL
4. 若历史本地库结构冲突严重，可重建本地开发库

### 启动顺序

```text
1. docker compose up -d mysql
2. 配置 .env
3. go run cmd/server/main.go
```

### 重置本地数据库

如果需要重置本地开发环境：

```bash
docker compose down -v
```

然后重新执行：

```bash
docker compose up -d mysql
```

## 测试

### 后端单元测试

```bash
go test ./...
```

### 集成测试

集成测试依赖 MySQL，并默认连接测试库 `stock_analysis_test`。
可通过环境变量覆盖：

```bash
TEST_DB_HOST=127.0.0.1
TEST_DB_PORT=3306
TEST_DB_USER=root
TEST_DB_PASSWORD=root123
TEST_DB_NAME=stock_analysis_test
```

运行方式示例：

```bash
go test -tags integration ./internal/repository/...
```

### 测试脚本

项目根目录下还提供了：

```bash
./test/run_tests.sh
```

该脚本默认执行常规 Go 测试，不会自动带上 `integration` 标签。
