# go-rpc Makefile - 构建、生成、测试入口

# ==================== 变量定义 ====================
MODULE_NAME    := github.com/desyang/go-rpc
GO             := go
PROTOC         := protoc
PROTOC_GEN_GO  := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc

# Go 版本要求
GO_VERSION     := 1.22

# 输出目录
BIN_DIR        := ./bin
GENERATED_DIR  := ./gen

# proto 文件路径
PROTO_DIRS     := ./api

# 所有 proto 文件
PROTO_FILES    := $(wildcard $(PROTO_DIRS)/*.proto)

# ==================== 伪目标 ====================
.PHONY: all setup proto proto-python proto-ts build clean test bench help

# ==================== 默认目标 ====================
all: setup proto build test

# ==================== 环境初始化 ====================
setup:
	@echo "=== 初始化开发环境 ==="
	@echo "检查 Go 版本..."
	$(GO) version
	@echo "安装 protoc 插件..."
	@if ! command -v $(PROTOC_GEN_GO) &> /dev/null; then \
		echo "安装 protoc-gen-go..."; \
		$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if ! command -v $(PROTOC_GEN_GO_GRPC) &> /dev/null; then \
		echo "安装 protoc-gen-go-grpc..."; \
		$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi
	@echo "安装依赖..."
	$(GO) mod download
	@echo "=== 环境初始化完成 ==="

# ==================== Protobuf 代码生成 ====================
proto: $(GENERATED_DIR)/go/api
	@echo "=== 生成 gRPC 代码 (Go) ==="
	$(PROTOC) --go_out=$(GENERATED_DIR)/go --go_opt=paths=source_relative \
		--go-grpc_out=$(GENERATED_DIR)/go --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

$(GENERATED_DIR)/go/api: $(PROTO_FILES)
	@mkdir -p $(GENERATED_DIR)/go/api

# ==================== Python 代码生成（示例脚本） ====================
proto-python:
	@echo "=== 生成 Python 客户端代码 ==="
	@if [ ! -d "./gen/python" ]; then mkdir -p ./gen/python; fi
	@# 需要安装 grpcio-tools: pip install grpcio-tools
	$(PROTOC) --python_out=./gen/python --grpc-python_out=./gen/python \
		$(PROTO_FILES)

# ==================== TypeScript 代码生成（示例脚本） ====================
proto-ts:
	@echo "=== 生成 TypeScript gRPC-Web 代码 ==="
	@if [ ! -d "./gen/ts" ]; then mkdir -p ./gen/ts; fi
	@# 需要安装 grpc_tools_node_protoc: npm install -g grpc_tools_node_protoc
	$(PROTOC) --grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:./gen/ts \
		$(PROTO_FILES)

# ==================== 构建 ====================
build:
	@echo "=== 构建 Go 服务端 ==="
	$(GO) build -o $(BIN_DIR)/rpc-server ./cmd/rpc-server/
	@echo "=== 构建 Go 调试客户端 ==="
	$(GO) build -o $(BIN_DIR)/rpc-client ./cmd/rpc-client/
	@echo "=== 构建 rpc-gen 代码生成工具 ==="
	$(GO) build -o $(BIN_DIR)/rpc-gen ./cmd/rpc-gen/
	@echo "=== 构建完成 ==="

# ==================== 清理 ====================
clean:
	@echo "=== 清理 ==="
	rm -rf $(BIN_DIR)
	rm -rf $(GENERATED_DIR)
	rm -rf ./gen/python ./gen/ts
	$(GO) clean

# ==================== 测试 ====================
test:
	@echo "=== 运行单元测试 ==="
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# ==================== 性能基准测试 ====================
bench:
	@echo "=== 运行基准测试 ==="
	$(GO) test -bench=. -benchmem -benchtime=10s ./...

# ==================== 检查 ====================
.PHONY: lint format
lint:
	@echo "=== 运行 golangci-lint ==="
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint 未安装，跳过 lint 检查"; \
		echo "安装方法: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format:
	@echo "=== 代码格式化 ==="
	@$(GO) fmt ./...

# ==================== 帮助 ====================
help:
	@echo ""
	@echo "go-rpc Makefile - 常用命令"
	@echo "========================"
	@echo ""
	@echo "  make setup      - 初始化开发环境"
	@echo "  make proto      - 生成 Go gRPC 桩代码"
	@echo "  make proto-python - 生成 Python 客户端代码"
	@echo "  make proto-ts   - 生成 TypeScript gRPC-Web 代码"
	@echo "  make build      - 构建所有二进制文件"
	@echo "  make test       - 运行单元测试"
	@echo "  make bench      - 运行性能基准测试"
	@echo "  make lint       - 代码 lint 检查"
	@echo "  make format     - 代码格式化"
	@echo "  make clean      - 清理构建产物"
	@echo ""
