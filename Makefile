PROGRAM_NAME := mac-guest-agent
VERSION := 1.1.0
BUILD_DIR := build
DIST_DIR := dist
BINARY := $(BUILD_DIR)/$(PROGRAM_NAME)

# 默认构建为当前架构
ARCH := $(shell uname -m)
ifeq ($(ARCH),arm64)
    GOARCH := arm64
else
    GOARCH := amd64
endif

.PHONY: all build clean install uninstall run test deps help build-amd64 build-arm64 build-all

# 默认目标
all: build

# 构建程序（当前架构）
build:
	@echo "构建 $(PROGRAM_NAME) v$(VERSION) ($(GOARCH))..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) go build \
		-ldflags "-X main.version=$(VERSION) -s -w" \
		-o $(BINARY) \
		main.go
	@echo "构建完成: $(BINARY)"

# 构建 AMD64 架构
build-amd64:
	@echo "构建 $(PROGRAM_NAME) v$(VERSION) (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -s -w" \
		-o $(BUILD_DIR)/$(PROGRAM_NAME)-darwin-amd64 \
		main.go
	@echo "AMD64 构建完成: $(BUILD_DIR)/$(PROGRAM_NAME)-darwin-amd64"

# 构建 ARM64 架构
build-arm64:
	@echo "构建 $(PROGRAM_NAME) v$(VERSION) (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags "-X main.version=$(VERSION) -s -w" \
		-o $(BUILD_DIR)/$(PROGRAM_NAME)-darwin-arm64 \
		main.go
	@echo "ARM64 构建完成: $(BUILD_DIR)/$(PROGRAM_NAME)-darwin-arm64"

# 构建多架构版本
build-all: build-amd64 build-arm64
	@echo "多架构构建完成"
	@ls -la $(BUILD_DIR)/$(PROGRAM_NAME)-darwin-*

# 生成校验和
checksums: build-all
	@echo "生成校验和..."
	@cd $(BUILD_DIR) && md5 $(PROGRAM_NAME)-darwin-amd64 > $(PROGRAM_NAME)-darwin-amd64.md5
	@cd $(BUILD_DIR) && md5 $(PROGRAM_NAME)-darwin-arm64 > $(PROGRAM_NAME)-darwin-arm64.md5
	@cd $(BUILD_DIR) && shasum -a 256 $(PROGRAM_NAME)-darwin-amd64 > $(PROGRAM_NAME)-darwin-amd64.sha256
	@cd $(BUILD_DIR) && shasum -a 256 $(PROGRAM_NAME)-darwin-arm64 > $(PROGRAM_NAME)-darwin-arm64.sha256
	@echo "校验和生成完成"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@echo "清理完成"

# 安装依赖
deps:
	@echo "安装依赖..."
	@go mod download
	@go mod tidy
	@echo "依赖安装完成"

# 运行程序（开发模式）
run: build
	@echo "运行程序..."
	@sudo $(BINARY) --verbose

# 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 代码格式化
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@golangci-lint run ./... || echo "请安装 golangci-lint: brew install golangci-lint"

# 创建发布包
dist: build-all checksums
	@echo "创建发布包..."
	@mkdir -p $(DIST_DIR)
	@cp $(BUILD_DIR)/$(PROGRAM_NAME)-darwin-* $(DIST_DIR)/
	@echo "发布包创建完成: $(DIST_DIR)/"

# 安装到系统
install: build
	@echo "安装到系统..."
	@sudo $(BINARY) --install

# 本地安装（直接使用本地构建的二进制文件）
install-local: build
	@echo "本地安装..."
	@sudo ./scripts/local_install.sh

# 从系统卸载
uninstall:
	@echo "从系统卸载..."
	@sudo /usr/local/bin/$(PROGRAM_NAME) --uninstall || echo "程序未安装或已卸载"

# 查看服务状态
status:
	@echo "查看服务状态..."
	@sudo launchctl list com.macos.guest-agent || echo "服务未运行"

# 查看日志
logs:
	@echo "查看日志..."
	@tail -f /var/log/mac-guest-agent.log

# 重启服务
restart:
	@echo "重启服务..."
	@sudo launchctl stop com.macos.guest-agent || true
	@sudo launchctl start com.macos.guest-agent

# 显示帮助
help:
	@echo "macOS Guest Agent 构建系统"
	@echo ""
	@echo "可用命令:"
	@echo "  build         - 构建程序（当前架构）"
	@echo "  build-amd64   - 构建 AMD64 架构"
	@echo "  build-arm64   - 构建 ARM64 架构"
	@echo "  build-all     - 构建所有架构"
	@echo "  checksums     - 生成校验和"
	@echo "  clean         - 清理构建文件"
	@echo "  deps          - 安装依赖"
	@echo "  run           - 运行程序（需要sudo）"
	@echo "  test          - 运行测试"
	@echo "  fmt           - 格式化代码"
	@echo "  lint          - 代码检查"
	@echo "  dist          - 创建发布包"
	@echo "  install       - 安装到系统"
	@echo "  install-local - 本地安装"
	@echo "  uninstall     - 从系统卸载"
	@echo "  status        - 查看服务状态"
	@echo "  logs          - 查看日志"
	@echo "  restart       - 重启服务"
	@echo "  help          - 显示此帮助" 