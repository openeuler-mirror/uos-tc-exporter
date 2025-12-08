// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package logger

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		logPath  string
		maxSize  int64
		maxAge   time.Duration
		expected fileLogConfig
	}{
		{
			name:    "valid config with file rotator",
			level:   "info",
			logPath: "/tmp/test.log",
			maxSize: 1024 * 1024, // 1MB
			maxAge:  time.Hour * 24,
			expected: fileLogConfig{
				level: "info",
			},
		},
		{
			name:    "valid config with debug level",
			level:   "debug",
			logPath: "/tmp/debug.log",
			maxSize: 512 * 1024, // 512KB
			maxAge:  time.Hour * 12,
			expected: fileLogConfig{
				level: "debug",
			},
		},
		{
			name:    "empty log path",
			level:   "info",
			logPath: "",
			maxSize: 1024 * 1024,
			maxAge:  time.Hour * 24,
			expected: fileLogConfig{
				level: "info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig(tt.level, tt.logPath, tt.maxSize, tt.maxAge)

			assert.Equal(t, tt.expected.level, config.level)
			if tt.logPath != "" {
				assert.NotNil(t, config.FileRotator)
				assert.Equal(t, tt.logPath, config.FileRotator.basePath)
				assert.Equal(t, tt.maxSize, config.FileRotator.maxSize)
				assert.Equal(t, tt.maxAge, config.FileRotator.maxAge)
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name          string
		config        fileLogConfig
		expectedLevel logrus.Level
		shouldSetFile bool
	}{
		{
			name: "debug level with file rotator",
			config: fileLogConfig{
				level:       "debug",
				FileRotator: NewFileRotator("/tmp/test_debug.log", 1024*1024, time.Hour*24),
			},
			expectedLevel: logrus.DebugLevel,
			shouldSetFile: true,
		},
		{
			name: "info level with file rotator",
			config: fileLogConfig{
				level:       "info",
				FileRotator: NewFileRotator("/tmp/test_info.log", 1024*1024, time.Hour*24),
			},
			expectedLevel: logrus.InfoLevel,
			shouldSetFile: true,
		},
		{
			name: "warn level with file rotator",
			config: fileLogConfig{
				level:       "warn",
				FileRotator: NewFileRotator("/tmp/test_warn.log", 1024*1024, time.Hour*24),
			},
			expectedLevel: logrus.WarnLevel,
			shouldSetFile: true,
		},
		{
			name: "error level with file rotator",
			config: fileLogConfig{
				level:       "error",
				FileRotator: NewFileRotator("/tmp/test_error.log", 1024*1024, time.Hour*24),
			},
			expectedLevel: logrus.ErrorLevel,
			shouldSetFile: true,
		},
		{
			name: "empty level (default to info)",
			config: fileLogConfig{
				level:       "",
				FileRotator: nil,
			},
			expectedLevel: logrus.InfoLevel,
			shouldSetFile: false,
		},
		{
			name: "unknown level (default to warn)",
			config: fileLogConfig{
				level:       "unknown",
				FileRotator: nil,
			},
			expectedLevel: logrus.WarnLevel,
			shouldSetFile: false,
		},
		{
			name: "case insensitive level",
			config: fileLogConfig{
				level:       "DEBUG",
				FileRotator: nil,
			},
			expectedLevel: logrus.DebugLevel,
			shouldSetFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始状态以便恢复
			originalLevel := logrus.GetLevel()
			originalOutput := logrus.StandardLogger().Out
			originalFormatter := logrus.StandardLogger().Formatter
			originalCallerSetting := logrus.StandardLogger().ReportCaller

			defer func() {
				logrus.SetLevel(originalLevel)
				logrus.SetOutput(originalOutput)
				logrus.SetFormatter(originalFormatter)
				logrus.SetReportCaller(originalCallerSetting)
			}()

			// 执行初始化
			Init(tt.config)

			// 验证日志级别
			assert.Equal(t, tt.expectedLevel, logrus.GetLevel())

			// 验证输出设置
			if tt.shouldSetFile {
				assert.NotEqual(t, originalOutput, logrus.StandardLogger().Out)
				assert.True(t, logrus.StandardLogger().ReportCaller)
			} else {
				assert.Equal(t, originalOutput, logrus.StandardLogger().Out)
			}

			// 验证调用者信息设置
			if tt.config.level == "debug" {
				assert.True(t, logrus.StandardLogger().ReportCaller)
			}
		})
	}
}

func TestInitDefaultLog(t *testing.T) {
	// 保存原始状态以便恢复
	originalLevel := logrus.GetLevel()
	originalOutput := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter

	defer func() {
		logrus.SetLevel(originalLevel)
		logrus.SetOutput(originalOutput)
		logrus.SetFormatter(originalFormatter)
	}()

	// 执行默认初始化
	InitDefaultLog()

	// 验证默认设置
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	assert.Equal(t, originalOutput, logrus.StandardLogger().Out)

	// 验证格式化器设置
	assert.NotNil(t, logrus.StandardLogger().Formatter)
}

func TestInit_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config fileLogConfig
	}{
		{
			name: "nil file rotator",
			config: fileLogConfig{
				level:       "info",
				FileRotator: nil,
			},
		},
		{
			name: "invalid log path for file rotator",
			config: fileLogConfig{
				level:       "info",
				FileRotator: NewFileRotator("/invalid/path/test.log", 1024*1024, time.Hour*24),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始状态以便恢复
			originalLevel := logrus.GetLevel()
			originalOutput := logrus.StandardLogger().Out

			defer func() {
				logrus.SetLevel(originalLevel)
				logrus.SetOutput(originalOutput)
			}()

			// 应该不会panic
			assert.NotPanics(t, func() {
				Init(tt.config)
			})

			// 验证基本设置
			assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel logrus.Level
	}{
		{"debug level", "debug", logrus.DebugLevel},
		{"info level", "info", logrus.InfoLevel},
		{"warn level", "warn", logrus.WarnLevel},
		{"error level", "error", logrus.ErrorLevel},
		{"case insensitive debug", "DEBUG", logrus.DebugLevel},
		{"case insensitive info", "INFO", logrus.InfoLevel},
		{"case insensitive warn", "WARN", logrus.WarnLevel},
		{"case insensitive error", "ERROR", logrus.ErrorLevel},
		{"mixed case", "DeBuG", logrus.DebugLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始状态以便恢复
			originalLevel := logrus.GetLevel()
			defer func() {
				logrus.SetLevel(originalLevel)
			}()

			config := fileLogConfig{
				level: tt.level,
			}

			Init(config)

			assert.Equal(t, tt.expectedLevel, logrus.GetLevel())
		})
	}
}

func TestUnknownLogLevel(t *testing.T) {
	// 保存原始状态以便恢复
	originalLevel := logrus.GetLevel()
	originalOutput := logrus.StandardLogger().Out

	defer func() {
		logrus.SetLevel(originalLevel)
		logrus.SetOutput(originalOutput)
	}()

	// 捕获标准输出以验证警告消息
	old := logrus.StandardLogger().Out
	r, w, _ := os.Pipe()
	logrus.SetOutput(w)

	config := fileLogConfig{
		level: "invalid_level",
	}

	Init(config)

	w.Close()
	logrus.SetOutput(old)

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// 验证使用了默认的warn级别
	assert.Equal(t, logrus.WarnLevel, logrus.GetLevel())

	// 验证警告消息包含相关信息
	assert.Contains(t, outputStr, "Unknown log level")
	assert.Contains(t, outputStr, "invalid_level")
	assert.Contains(t, outputStr, "using default level: warn")
}

func TestFileRotatorIntegration(t *testing.T) {
	tempDir := t.TempDir()
	logPath := tempDir + "/test.log"

	tests := []struct {
		name    string
		level   string
		logPath string
	}{
		{
			name:    "file rotator with debug level",
			level:   "debug",
			logPath: logPath,
		},
		{
			name:    "file rotator with info level",
			level:   "info",
			logPath: logPath + ".info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始状态以便恢复
			originalLevel := logrus.GetLevel()
			originalOutput := logrus.StandardLogger().Out

			defer func() {
				logrus.SetLevel(originalLevel)
				logrus.SetOutput(originalOutput)
			}()

			config := fileLogConfig{
				level:       tt.level,
				FileRotator: NewFileRotator(tt.logPath, 1024*1024, time.Hour*24),
			}

			Init(config)

			// 验证文件存在
			_, err := os.Stat(tt.logPath)
			assert.NoError(t, err)

			// 验证日志级别设置正确
			switch strings.ToLower(tt.level) {
			case "debug":
				assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
			case "info":
				assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
			}

			// 验证调用者信息设置（debug级别应该启用）
			if strings.ToLower(tt.level) == "debug" {
				assert.True(t, logrus.StandardLogger().ReportCaller)
			}
		})
	}
}

// 基准测试
func BenchmarkInit(b *testing.B) {
	tempDir := b.TempDir()
	logPath := tempDir + "/benchmark.log"

	config := fileLogConfig{
		level:       "info",
		FileRotator: NewFileRotator(logPath, 1024*1024, time.Hour*24),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Init(config)
	}
}

func BenchmarkInitDefaultLog(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InitDefaultLog()
	}
}

func BenchmarkNewConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewConfig("info", "/tmp/test.log", 1024*1024, time.Hour*24)
	}
}
