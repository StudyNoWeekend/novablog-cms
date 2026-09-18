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

# 构建容器镜像（单镜像：nginx + 后端 + 管理后台前端）
docker-build:
	docker build -f ../deploy/Dockerfile --build-arg VERSION=dev -t novablog:latest ..

# 容器部署（推荐改用部署脚本，可指定端口/数据库/挂载目录：cd ../deploy && ./deploy.sh --help）
docker-up:
	cd ../deploy && ./deploy.sh --build --yes

docker-down:
	cd ../deploy && ./deploy.sh --down