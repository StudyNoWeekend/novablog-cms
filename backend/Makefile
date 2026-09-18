.PHONY: build run fmt vet lint test clean docker-build docker-up docker-down

# 编译
build:
	go build -o bin/novablog ./cmd/api

# 运行
run:
	go run ./cmd/api

# 格式化代码
fmt:
	go fmt ./...

# 静态检查
vet:
	go vet ./...

# 代码规范检查
lint:
	golangci-lint run ./...

# 运行测试
test:
	go test -v -race -coverprofile=coverage.out ./...

# 清理编译产物
clean:
	rm -rf bin/
	rm -f coverage.out

# 构建容器镜像（后端 + nginx 入口）
docker-build:
	docker build -f ../deploy/Dockerfile -t novablog-backend:latest ..
	docker build -f ../deploy/nginx.Dockerfile -t novablog-nginx:latest ..

# 容器部署（首次使用：cd ../deploy && cp .env.example .env 并修改密码/域名）
docker-up:
	cd ../deploy && docker compose up -d --build

docker-down:
	cd ../deploy && docker compose down