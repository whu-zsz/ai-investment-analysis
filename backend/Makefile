.PHONY: help run build test clean swagger init-db db-up db-down test-integration

help:
	@echo "可用命令:"
	@echo "  make db-up             - 启动本地 MySQL 容器"
	@echo "  make db-down           - 停止并删除本地数据库容器与数据卷"
	@echo "  make run               - 运行服务"
	@echo "  make build             - 编译项目"
	@echo "  make test              - 运行测试"
	@echo "  make test-integration  - 运行 repository 集成测试"
	@echo "  make clean             - 清理构建文件"
	@echo "  make swagger           - 生成 Swagger 文档"
	@echo "  make init-db           - 初始化开发/测试数据库"

run:
	@echo "启动服务..."
	go run cmd/server/main.go

build:
	@echo "编译项目..."
	go build -o bin/server cmd/server/main.go

test:
	@echo "运行测试..."
	go test -v ./...

test-integration:
	@echo "运行 repository 集成测试..."
	go test -tags integration -v ./internal/repository/...

clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -rf uploads/*
	@touch uploads/.gitkeep

swagger:
	@echo "生成Swagger文档..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/server/main.go -o docs

init-db:
	@echo "初始化开发/测试数据库..."
	@echo "请确保本地 MySQL 已安装并运行，或优先使用 make db-up"
	@echo "请输入MySQL root密码:"
	@mysql -u root -p < scripts/init_db.sql

db-up:
	@echo "启动本地 MySQL 容器..."
	docker compose up -d mysql

db-down:
	@echo "停止并删除本地数据库容器与数据卷..."
	docker compose down -v

deps:
	@echo "安装依赖..."
	go mod download
	go mod tidy

fmt:
	@echo "格式化代码..."
	@which gofmt > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	go fmt ./...
	goimports -w .

lint:
	@echo "代码检查..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

docker-build:
	@echo "构建Docker镜像..."
	docker build -t stock-analysis-backend:latest .

docker-run:
	@echo "运行Docker容器..."
	docker run -p 8080:8080 --env-file .env stock-analysis-backend:latest

dev:
	@echo "启动开发模式..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air
