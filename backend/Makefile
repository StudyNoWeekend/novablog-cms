.PHONY: build run fmt vet lint test clean

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