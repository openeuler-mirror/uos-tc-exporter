# SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
# SPDX-License-Identifier: MIT

# 项目信息
PROJECT_NAME := uos-tc-exporter
BINARY_NAME := uos_tc_exporter
VERSION := $(shell cat version/version.go | grep 'Version.*=' | head -1 | sed 's/.*Version.*=.*"\(.*\)"/\1/')
REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BuildTime=$(BUILD_TIME)"

# Go 相关变量
GO := go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOBUILD := $(GO) build
GOTEST := $(GO) test
GOCLEAN := $(GO) clean
GOMOD := $(GO) mod
GOGET := $(GO) get

# 构建目录
BUILD_DIR := build
DIST_DIR := dist
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)
TARBALL := $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz

# 默认目标
.PHONY: all
all: clean build

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  dist       - Create distribution tarball"
	@echo "  install    - Install binary to system"
	@echo "  uninstall  - Remove installed binary"
	@echo "  deps       - Download dependencies"
	@echo "  tidy       - Tidy go.mod and go.sum"
	@echo "  fmt        - Format source code"
	@echo "  lint       - Run linter"
	@echo "  vet        - Run go vet"
	@echo "  race       - Run tests with race detection"
	@echo "  coverage   - Run tests with coverage"
	@echo "  docker     - Build Docker image"
	@echo "  help       - Show this help message"

# 创建构建目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# 构建二进制文件
.PHONY: build
build: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@echo "  OS: $(GOOS)"
	@echo "  Arch: $(GOARCH)"
	@echo "  Revision: $(REVISION)"
	@echo "  Build Time: $(BUILD_TIME)"
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) .
	@echo "Build completed: $(BINARY_PATH)"

# 交叉编译
.PHONY: build-linux
build-linux: GOOS=linux
build-linux: build

.PHONY: build-windows
build-windows: GOOS=windows
build-windows: build

.PHONY: build-darwin
build-darwin: GOOS=darwin
build-darwin: build

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 运行竞态检测
.PHONY: race
race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting source code..."
	$(GO) fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# 代码检查（如果有 golangci-lint）
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v -t -d ./...

# 整理依赖
.PHONY: tidy
tidy:
	@echo "Tidying go.mod and go.sum..."
	$(GOMOD) tidy
	$(GOMOD) verify

# 创建发布包
.PHONY: dist
dist: build $(DIST_DIR)
	@echo "Creating distribution package..."
	@mkdir -p $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)
	@cp $(BINARY_PATH) $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp README.md $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp README.en.md $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp LICENSE $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp -r config $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp -r docs $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cd $(DIST_DIR) && tar -czf $(PROJECT_NAME)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz $(PROJECT_NAME)-$(VERSION)
	@rm -rf $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)
	@echo "Distribution package created: $(TARBALL)"

# 安装到系统
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_PATH) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation completed"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation completed"

# Docker 构建
.PHONY: docker
docker:
	@echo "Building Docker image..."
	docker build -t $(PROJECT_NAME):$(VERSION) .
	docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest
	@echo "Docker image built: $(PROJECT_NAME):$(VERSION)"

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out coverage.html
	@$(GOCLEAN)
	@echo "Clean completed"

# 完全清理（包括依赖）
.PHONY: clean-all
clean-all: clean
	@echo "Cleaning all dependencies..."
	@rm -rf vendor/
	@rm -f go.sum
	@echo "Complete clean completed"

# 显示版本信息
.PHONY: version
version:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Revision: $(REVISION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "OS/Arch: $(GOOS)/$(GOARCH)"

# 开发模式：监听文件变化并自动重新构建
.PHONY: dev
dev:
	@echo "Starting development mode..."
	@echo "Watching for file changes..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not found, install with: go install github.com/cosmtrek/air@latest"; \
		echo "Or manually run: make build && ./$(BINARY_PATH)"; \
	fi

# 检查依赖更新
.PHONY: deps-check
deps-check:
	@echo "Checking for dependency updates..."
	@if command -v go-mod-outdated >/dev/null 2>&1; then \
		go-mod-outdated -update -direct; \
	else \
		echo "go-mod-outdated not found, install with: go install github.com/psampaz/go-mod-outdated@latest"; \
	fi

# 安全扫描
.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found, install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 性能基准测试
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# 显示构建信息
.PHONY: info
info:
	@echo "Build Information:"
	@echo "  Project: $(PROJECT_NAME)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Revision: $(REVISION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(shell go version)"
	@echo "  OS/Arch: $(GOOS)/$(GOARCH)"
	@echo "  Build Dir: $(BUILD_DIR)"
	@echo "  Dist Dir: $(DIST_DIR)"
	@echo "  Binary Path: $(BINARY_PATH)"
	@echo "  Tarball: $(TARBALL)"
