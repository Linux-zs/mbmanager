.PHONY: build run clean docker-build docker-up docker-down test

# 构建
build:
	go build -o bin/dbbak cmd/server/main.go

# 运行
run:
	go run cmd/server/main.go

# 清理
clean:
	rm -rf bin/
	rm -rf data/

# Docker构建
docker-build:
	docker build -f docker/Dockerfile -t dbbak:latest .

# Docker启动
docker-up:
	cd docker && docker-compose up -d

# Docker停止
docker-down:
	cd docker && docker-compose down

# 测试
test:
	go test -v ./...

# 安装依赖
deps:
	go mod tidy
	go mod download

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 帮助
help:
	@echo "可用命令："
	@echo "  make build        - 构建项目"
	@echo "  make run          - 运行项目"
	@echo "  make clean        - 清理构建文件"
	@echo "  make docker-build - 构建Docker镜像"
	@echo "  make docker-up    - 启动Docker容器"
	@echo "  make docker-down  - 停止Docker容器"
	@echo "  make test         - 运行测试"
	@echo "  make deps         - 安装依赖"
	@echo "  make fmt          - 格式化代码"
	@echo "  make lint         - 代码检查"
