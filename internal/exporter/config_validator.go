// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/pkg/errors"
	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	validator *validator.Validate
}

// NewConfigValidator 创建新的配置验证器
func NewConfigValidator() *ConfigValidator {
	v := validator.New()
	cv := &ConfigValidator{
		validator: v,
	}

	// 注册自定义验证规则
	cv.registerCustomValidations()

	return cv
}

// registerCustomValidations 注册自定义验证规则
func (cv *ConfigValidator) registerCustomValidations() {
	// 验证IP地址
	cv.validator.RegisterValidation("ip", func(fl validator.FieldLevel) bool {
		ip := fl.Field().String()
		if ip == "" {
			return true // 空值由其他验证器处理
		}
		return net.ParseIP(ip) != nil
	})

	// 验证端口号
	cv.validator.RegisterValidation("port", func(fl validator.FieldLevel) bool {
		port := fl.Field().Int()
		return port > 0 && port <= 65535
	})

	// 验证时间格式
	cv.validator.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		durationStr := fl.Field().String()
		if durationStr == "" {
			return true
		}
		_, err := time.ParseDuration(durationStr)
		return err == nil
	})

	// 验证日志级别
	cv.validator.RegisterValidation("loglevel", func(fl validator.FieldLevel) bool {
		level := strings.ToLower(fl.Field().String())
		validLevels := map[string]bool{
			"debug":   true,
			"info":    true,
			"warn":    true,
			"warning": true,
			"error":   true,
			"fatal":   true,
			"panic":   true,
		}
		return validLevels[level]
	})

	// 验证文件路径
	cv.validator.RegisterValidation("filepath", func(fl validator.FieldLevel) bool {
		path := fl.Field().String()
		if path == "" {
			return true
		}
		// 简单的路径格式检查
		return !strings.Contains(path, "..") && !strings.Contains(path, "//")
	})

	// 验证指标路径
	cv.validator.RegisterValidation("metricspath", func(fl validator.FieldLevel) bool {
		path := fl.Field().String()
		if path == "" {
			return false
		}
		// 必须以斜杠开头
		if !strings.HasPrefix(path, "/") {
			return false
		}
		// 不能包含特殊字符
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\-/]+$`, path)
		return matched
	})
}

// ValidateConfig 验证配置
func (cv *ConfigValidator) ValidateConfig(config *Config) error {
	if config == nil {
		return errors.New(errors.ErrCodeConfigValidation, "config is nil")
	}

	// 基本验证
	if err := cv.validator.Struct(config); err != nil {
		return cv.wrapValidationError(err)
	}

	// 高级验证
	if err := cv.validateAdvanced(config); err != nil {
		return err
	}

	return nil
}

// validateAdvanced 高级验证
func (cv *ConfigValidator) validateAdvanced(config *Config) error {
	// 验证监听地址和端口组合
	if config.Address == "0.0.0.0" && config.Port < 1024 {
		return errors.NewWithContext(
			errors.ErrCodeConfigValidation,
			"binding to privileged port with wildcard address is not recommended",
			map[string]interface{}{
				"address": config.Address,
				"port":    config.Port,
			},
		)
	}

	// 验证日志配置
	if err := cv.validateLogConfig(&config.Logging); err != nil {
		return err
	}

	// 验证服务器配置
	if err := cv.validateServerConfig(&config.Server); err != nil {
		return err
	}

	return nil
}

// validateLogConfig 验证日志配置
func (cv *ConfigValidator) validateLogConfig(logConfig *logger.Config) error {
	if logConfig.MaxSize != "" {
		// 验证文件大小格式 (e.g., "10MB", "1GB")
		sizeRegex := regexp.MustCompile(`^(\d+)([KMG]B)?$`)
		if !sizeRegex.MatchString(logConfig.MaxSize) {
			return errors.NewWithContext(
				errors.ErrCodeConfigValidation,
				"invalid log max size format",
				map[string]interface{}{
					"max_size": logConfig.MaxSize,
					"expected": "e.g., 10MB, 1GB",
				},
			)
		}
	}

	if logConfig.MaxAge < 0 {
		return errors.NewWithContext(
			errors.ErrCodeConfigValidation,
			"log max age cannot be negative",
			map[string]interface{}{
				"max_age": logConfig.MaxAge,
			},
		)
	}

	return nil
}

// validateServerConfig 验证服务器配置
func (cv *ConfigValidator) validateServerConfig(serverConfig *ServerConfig) error {
	if serverConfig.ShutdownTimeout != 0 {
		duration := serverConfig.ShutdownTimeout

		if duration < 5*time.Second {
			return errors.NewWithContext(
				errors.ErrCodeConfigValidation,
				"shutdown timeout is too short",
				map[string]interface{}{
					"shutdown_timeout": serverConfig.ShutdownTimeout,
					"minimum":          "5s",
				},
			)
		}

		if duration > 5*time.Minute {
			return errors.NewWithContext(
				errors.ErrCodeConfigValidation,
				"shutdown timeout is too long",
				map[string]interface{}{
					"shutdown_timeout": serverConfig.ShutdownTimeout,
					"maximum":          "5m",
				},
			)
		}
	}

	return nil
}

// wrapValidationError 包装验证错误
func (cv *ConfigValidator) wrapValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorDetails []map[string]interface{}

		for _, fieldError := range validationErrors {
			detail := map[string]interface{}{
				"field":   fieldError.Field(),
				"value":   fieldError.Value(),
				"tag":     fieldError.Tag(),
				"param":   fieldError.Param(),
				"message": cv.getValidationMessage(fieldError),
			}
			errorDetails = append(errorDetails, detail)
		}

		return errors.NewWithContext(
			errors.ErrCodeConfigValidation,
			"configuration validation failed",
			map[string]interface{}{
				"validation_errors": errorDetails,
				"total_errors":      len(validationErrors),
			},
		)
	}

	return errors.Wrap(err, errors.ErrCodeConfigValidation, "configuration validation failed")
}

// getValidationMessage 获取验证错误消息
func (cv *ConfigValidator) getValidationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", fieldError.Field())
	case "ip":
		return fmt.Sprintf("field '%s' must be a valid IP address", fieldError.Field())
	case "port":
		return fmt.Sprintf("field '%s' must be a valid port number (1-65535)", fieldError.Field())
	case "duration":
		return fmt.Sprintf("field '%s' must be a valid duration string (e.g., 30s, 1m)", fieldError.Field())
	case "loglevel":
		return fmt.Sprintf("field '%s' must be a valid log level (debug, info, warn, error, fatal, panic)", fieldError.Field())
	case "filepath":
		return fmt.Sprintf("field '%s' must be a valid file path", fieldError.Field())
	case "metricspath":
		return fmt.Sprintf("field '%s' must be a valid metrics path starting with '/' and containing only alphanumeric characters, hyphens, and underscores", fieldError.Field())
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s", fieldError.Field(), fieldError.Param())
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s", fieldError.Field(), fieldError.Param())
	default:
		return fmt.Sprintf("field '%s' failed validation with tag '%s'", fieldError.Field(), fieldError.Tag())
	}
}

// ValidateField 验证单个字段
func (cv *ConfigValidator) ValidateField(value interface{}, tag string) error {
	return cv.validator.Var(value, tag)
}

// GetValidationRules 获取验证规则描述
func (cv *ConfigValidator) GetValidationRules() map[string]string {
	return map[string]string{
		"required":    "Field is required",
		"ip":          "Must be a valid IP address",
		"port":        "Must be a valid port number (1-65535)",
		"duration":    "Must be a valid duration string (e.g., 30s, 1m)",
		"loglevel":    "Must be a valid log level (debug, info, warn, error, fatal, panic)",
		"filepath":    "Must be a valid file path",
		"metricspath": "Must be a valid metrics path starting with '/'",
		"min":         "Minimum value constraint",
		"max":         "Maximum value constraint",
	}
}

// ConfigSchema 配置模式定义
type ConfigSchema struct {
	Fields map[string]FieldSchema
}

// FieldSchema 字段模式定义
type FieldSchema struct {
	Type        string
	Required    bool
	Default     interface{}
	Description string
	Constraints map[string]interface{}
}

// GetSchema 获取配置模式
func (cv *ConfigValidator) GetSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: map[string]FieldSchema{
			"Address": {
				Type:        "string",
				Required:    true,
				Default:     "127.0.0.1",
				Description: "Server listening address",
				Constraints: map[string]interface{}{
					"validation": "ip",
				},
			},
			"Port": {
				Type:        "integer",
				Required:    true,
				Default:     9062,
				Description: "Server listening port",
				Constraints: map[string]interface{}{
					"min": 1,
					"max": 65535,
				},
			},
			"MetricsPath": {
				Type:        "string",
				Required:    true,
				Default:     "/metrics",
				Description: "Metrics endpoint path",
				Constraints: map[string]interface{}{
					"validation": "metricspath",
				},
			},
			"Logging.Level": {
				Type:        "string",
				Required:    false,
				Default:     "info",
				Description: "Log level",
				Constraints: map[string]interface{}{
					"validation": "loglevel",
				},
			},
			"Server.ShutdownTimeout": {
				Type:        "string",
				Required:    false,
				Default:     "30s",
				Description: "Graceful shutdown timeout",
				Constraints: map[string]interface{}{
					"validation": "duration",
				},
			},
		},
	}
}

// ParseDuration 解析持续时间字符串
func (cv *ConfigValidator) ParseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, errors.WrapWithContext(
			err,
			errors.ErrCodeConfigValidation,
			"failed to parse duration",
			map[string]interface{}{
				"duration": durationStr,
			},
		)
	}

	return duration, nil
}

// ParseFileSize 解析文件大小字符串
func (cv *ConfigValidator) ParseFileSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	// 解析大小字符串 (e.g., "10MB", "1GB")
	re := regexp.MustCompile(`^(\d+)([KMG]B)?$`)
	matches := re.FindStringSubmatch(sizeStr)

	if len(matches) != 3 {
		return 0, errors.NewWithContext(
			errors.ErrCodeConfigValidation,
			"invalid file size format",
			map[string]interface{}{
				"size":    sizeStr,
				"pattern": "e.g., 10MB, 1GB",
			},
		)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, errors.WrapWithContext(
			err,
			errors.ErrCodeConfigValidation,
			"failed to parse file size value",
			map[string]interface{}{
				"size": sizeStr,
			},
		)
	}

	unit := matches[2]
	switch unit {
	case "KB":
		return value * 1024, nil
	case "MB":
		return value * 1024 * 1024, nil
	case "GB":
		return value * 1024 * 1024 * 1024, nil
	default: // 默认字节
		return value, nil
	}
}
