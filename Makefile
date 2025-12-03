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
GOSEC := gosec

# 构建目录
BUILD_DIR := bin
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)

# 默认目标
.PHONY: all
all: lint test gosec build

# 完整检查目标（包含安全扫描）
.PHONY: check
check: lint test gosec
	@echo "All checks completed successfully"

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run code quality checks"
	@echo "  gosec         - Run security scan with gosec"
	@echo "  gosec-detail  - Run detailed security scan with gosec"
	@echo "  gosec-html    - Run security scan and generate HTML report"
	@echo "  clean         - Clean build artifacts"
	@echo "  help          - Show this help message"

# 创建构建目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

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

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -covermode=atomic -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	@echo "Coverage report generated: coverage.out"
	@echo "HTML coverage report: coverage.html (run 'make coverage-html')"

# 生成HTML覆盖率报告
.PHONY: coverage-html
coverage-html: test-coverage
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"

# 运行代码质量检查
.PHONY: lint
lint:
	@echo "Running code quality checks..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2; \
	fi
	golangci-lint run ./...
	@echo "Code quality checks completed"

# 运行安全扫描
.PHONY: gosec
gosec:
	@echo "Running security scan with gosec..."
	@if ! command -v $(GOSEC) >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GO) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	$(GOSEC) ./...
	@echo "Security scan completed"
	
# 运行详细的安全扫描（包含更多信息）
.PHONY: gosec-detail
gosec-detail:
	@echo "Running detailed security scan with gosec..."
	@if ! command -v $(GOSEC) >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GO) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	$(GOSEC) -fmt=json -out=gosec-results.json ./...
	@echo "Detailed security scan completed. Results saved to gosec-results.json"

# 运行安全扫描并生成HTML报告
.PHONY: gosec-html
gosec-html:
	@echo "Running security scan with HTML report..."
	@if ! command -v $(GOSEC) >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		$(GO) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	$(GOSEC) -fmt=html -out=gosec-report.html ./...
	@echo "Security scan with HTML report completed. Report saved to gosec-report.html"

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "Clean completed"
