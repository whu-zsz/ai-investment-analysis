.PHONY: help run build test clean swagger init-db

# 默认目标
help:
	@echo "可用命令:"
	@echo "  make run        - 运行服务"
	@echo "  make build      - 编译项目"
	@echo "  make test       - 运行测试"
	@echo "  make clean      - 清理构建文件"
	@echo "  make swagger    - 生成Swagger文档"
	@echo "  make init-db    - 初始化数据库"

# 运行服务
run:
	@echo "启动服务..."
	go run cmd/server/main.go

# 编译项目
build:
	@echo "编译项目..."
	go build -o bin/server cmd/server/main.go

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -rf uploads/*
	@touch uploads/.gitkeep

# 生成Swagger文档
swagger:
	@echo "生成Swagger文档..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/server/main.go -o docs

# 初始化数据库
init-db:
	@echo "初始化数据库..."
	@echo "请确保MySQL已安装并运行"
	@echo "请输入MySQL root密码:"
	@mysql -u root -p < scripts/init_db.sql

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download
	go mod tidy

# 代码格式化
fmt:
	@echo "格式化代码..."
	@which gofmt > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	@echo "代码检查..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

# Docker构建
docker-build:
	@echo "构建Docker镜像..."
	docker build -t stock-analysis-backend:latest .

# Docker运行
docker-run:
	@echo "运行Docker容器..."
	docker run -p 8080:8080 --env-file .env stock-analysis-backend:latest

# 开发模式（热重载）
dev:
	@echo "启动开发模式..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air
