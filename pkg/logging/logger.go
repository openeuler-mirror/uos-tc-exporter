// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package logger

import (
	"strings"
	"time"

	formatter "gitee.com/weidongkl/logrus-formatter"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Level   string        `yaml:"level"`
	LogPath string        `yaml:"log_path"`
	MaxSize string        `yaml:"max_size"`
	MaxAge  time.Duration `yaml:"max_age"`
}

type fileLogConfig struct {
	FileRotator *FileRotator
	level       string
}

func NewConfig(level, logPath string, maxSize int64, maxAge time.Duration) fileLogConfig {
	return fileLogConfig{
		level:       level,
		FileRotator: NewFileRotator(logPath, maxSize, maxAge),
	}
}

func Init(config fileLogConfig) {
	// 设置日志格式
	logrus.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
		HideKeys:        true,
	})

	// 设置日志输出
	if config.FileRotator == nil {
		logrus.SetOutput(logrus.StandardLogger().Out)
	} else {
		logrus.SetReportCaller(true)
		logrus.SetOutput(config.FileRotator)
	}

	// 设置日志级别
	level := strings.ToLower(config.level)
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true) // 调试级别显示调用者信息
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "":
		// 默认级别为info
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Info("Using default log level: info")
	default:
		logrus.SetLevel(logrus.WarnLevel)
		logrus.Warnf("Unknown log level: %s, using default level: warn", level)
		logrus.Warn("Supported levels are: debug, info, warn, error")
	}

	// 设置性能优化选项
	logrus.SetNoLock() // 在单线程环境中提高性能

	logrus.Infof("Logger initialized with level: %s", logrus.GetLevel())
}

func InitDefaultLog() {
	Init(fileLogConfig{
		level: "info",
	})
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
}
