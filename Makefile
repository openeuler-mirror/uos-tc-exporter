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

# 构建目录
BUILD_DIR := bin
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)

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
	@echo "  help       - Show this help message"

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
	$(GOTEST) -v ./...

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "Clean completed"
