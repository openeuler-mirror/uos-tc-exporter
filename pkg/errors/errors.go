// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorCode 定义错误码类型
type ErrorCode int

const (
	// 系统级错误码 (1000-1999)
	ErrCodeSystem   ErrorCode = 1000
	ErrCodeConfig   ErrorCode = 1001
	ErrCodeNetwork  ErrorCode = 1002
	ErrCodeDatabase ErrorCode = 1003

	// 服务级错误码 (2000-2999)
	ErrCodeServer         ErrorCode = 2000
	ErrCodeServerSetup    ErrorCode = 2001
	ErrCodeServerRun      ErrorCode = 2002
	ErrCodeServerShutdown ErrorCode = 2003

	// 中间件错误码 (3000-3999)
	ErrCodeMiddleware ErrorCode = 3000
	ErrCodeRateLimit  ErrorCode = 3001
	ErrCodeAuth       ErrorCode = 3002

	// 业务逻辑错误码 (4000-4999)
	ErrCodeBusiness         ErrorCode = 4000
	ErrCodeMetrics          ErrorCode = 4001
	ErrCodeLandingPage      ErrorCode = 4002
	ErrCodeMetricsCollect   ErrorCode = 4003
	ErrCodeConfigValidation ErrorCode = 4004
	ErrCodeNetlinkOperation ErrorCode = 4005

	// TC 操作错误码 (5000-5999)
	ErrCodeTCOperation    ErrorCode = 5000
	ErrCodeQdiscOperation ErrorCode = 5001
	ErrCodeClassOperation ErrorCode = 5002
)

// Error 自定义错误结构
type Error struct {
	Code        ErrorCode
	Message     string
	Err         error
	Context     map[string]any
	Timestamp   time.Time
	Stack       []string
	CallerFile  string
	CallerLine  int
	CallerFunc  string
	Severity    string
	IsTemporary bool
}

// New 创建新的错误
func New(code ErrorCode, message string) *Error {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	return &Error{
		Code:        code,
		Message:     message,
		Context:     make(map[string]any),
		Timestamp:   time.Now(),
		CallerFile:  file,
		CallerLine:  line,
		CallerFunc:  fn.Name(),
		Severity:    getSeverityByCode(code),
		IsTemporary: isTemporaryByCode(code),
	}
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}

	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	var customErr *Error
	if e, ok := err.(*Error); ok {
		customErr = e
		// 更新包装信息
		customErr.Message = message
		customErr.Code = code
	} else {
		customErr = &Error{
			Code:        code,
			Message:     message,
			Err:         err,
			Context:     make(map[string]any),
			Timestamp:   time.Now(),
			CallerFile:  file,
			CallerLine:  line,
			CallerFunc:  fn.Name(),
			Severity:    getSeverityByCode(code),
			IsTemporary: isTemporaryByCode(code),
		}
	}

	return customErr
}

// WithContext 添加错误上下文
func (e *Error) WithContext(key string, value any) *Error {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithError 设置原始错误
func (e *Error) WithError(err error) *Error {
	e.Err = err
	return e
}

// Error 实现error接口
func (e *Error) Error() string {
	var builder strings.Builder

	if e.Code != 0 {
		builder.WriteString(fmt.Sprintf("[%d]", e.Code))
	}
	if e.Message != "" {
		builder.WriteString(e.Message)
	}

	if e.Err != nil {
		builder.WriteString(fmt.Sprintf(" (caused by: %v)", e.Err))
	}

	if len(e.Context) > 0 {
		contextJSON, _ := json.Marshal(e.Context)
		builder.WriteString(fmt.Sprintf(" [context: %s]", string(contextJSON)))
	}

	if e.CallerFile != "" {
		builder.WriteString(fmt.Sprintf(" [at %s:%d %s]", e.CallerFile, e.CallerLine, e.CallerFunc))
	}

	return builder.String()
}

// JSON 返回错误的JSON表示
func (e *Error) JSON() string {
	type ErrorJSON struct {
		Code        ErrorCode     `json:"code"`
		Message     string        `json:"message"`
		Cause       string        `json:"cause,omitempty"`
		Context     map[string]any `json:"context,omitempty"`
		Timestamp   time.Time     `json:"timestamp"`
		Severity    string        `json:"severity"`
		IsTemporary bool          `json:"is_temporary"`
		Caller      string        `json:"caller,omitempty"`
	}

	cause := ""
	if e.Err != nil {
		cause = e.Err.Error()
	}

	caller := ""
	if e.CallerFile != "" {
		caller = fmt.Sprintf("%s:%d %s", e.CallerFile, e.CallerLine, e.CallerFunc)
	}

	errorJSON := ErrorJSON{
		Code:        e.Code,
		Message:     e.Message,
		Cause:       cause,
		Context:     e.Context,
		Timestamp:   e.Timestamp,
		Severity:    e.Severity,
		IsTemporary: e.IsTemporary,
		Caller:      caller,
	}

	jsonBytes, _ := json.MarshalIndent(errorJSON, "", "  ")
	return string(jsonBytes)
}

// Unwrap 实现errors.Unwrap接口
func (e *Error) Unwrap() error {
	return e.Err
}

// GetCode 获取错误码
func (e *Error) GetCode() ErrorCode {
	return e.Code
}

// GetContext 获取错误上下文
func (e *Error) GetContext() map[string]any {
	return e.Context
}

// IsErrorCode 检查错误是否为指定错误码
func IsErrorCode(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// GetErrorCode 获取错误码，如果不是自定义错误则返回0
func GetErrorCode(err error) ErrorCode {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return 0
}

// GetErrorContext 获取错误上下文
func GetErrorContext(err error) map[string]any {
	if e, ok := err.(*Error); ok {
		return e.Context
	}
	return nil
}

// WrapWithContext 包装错误并添加上下文
func WrapWithContext(err error, code ErrorCode, message string, ctx map[string]any) *Error {
	customErr := Wrap(err, code, message)
	if customErr == nil {
		return nil
	}

	for k, v := range ctx {
		customErr.WithContext(k, v)
	}
	return customErr
}

// NewWithContext 创建新错误并添加上下文
func NewWithContext(code ErrorCode, message string, ctx map[string]any) *Error {
	err := New(code, message)
	for k, v := range ctx {
		err.WithContext(k, v)
	}
	return err
}

// IsTemporaryError 检查是否为临时错误（可重试）
func IsTemporaryError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.IsTemporary
	}
	code := GetErrorCode(err)
	// 网络错误、限流错误等通常是临时的
	return code == ErrCodeNetwork || code == ErrCodeRateLimit
}

// IsPermanentError 检查是否为永久错误（不可重试）
func IsPermanentError(err error) bool {
	return !IsTemporaryError(err)
}

// GetErrorSeverity 获取错误严重程度
func GetErrorSeverity(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Severity
	}
	code := GetErrorCode(err)
	return getSeverityByCode(code)
}

// 根据错误码获取严重程度
func getSeverityByCode(code ErrorCode) string {
	switch {
	case code >= 1000 && code < 2000:
		return "critical" // 系统级错误
	case code >= 2000 && code < 3000:
		return "high" // 服务级错误
	case code >= 3000 && code < 4000:
		return "medium" // 中间件错误
	case code >= 4000 && code < 6000:
		return "low" // 业务逻辑错误
	default:
		return "unknown"
	}
}

// 根据错误码判断是否为临时错误
func isTemporaryByCode(code ErrorCode) bool {
	// 网络错误、限流错误等通常是临时的
	return code == ErrCodeNetwork || code == ErrCodeRateLimit
}

// ShouldRetry 判断是否应该重试
func ShouldRetry(err error, maxRetries int, currentRetry int) bool {
	if currentRetry >= maxRetries {
		return false
	}
	return IsTemporaryError(err)
}

// GetRetryDelay 获取重试延迟时间（指数退避）
func GetRetryDelay(err error, currentRetry int) time.Duration {
	baseDelay := time.Second
	maxDelay := 30 * time.Second

	delay := baseDelay * time.Duration(1<<uint(currentRetry))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// ErrorStack 获取错误堆栈信息
func ErrorStack(err error) []string {
	var stack []string
	current := err

	for current != nil {
		if e, ok := current.(*Error); ok {
			stack = append(stack, e.JSON())
		} else {
			stack = append(stack, current.Error())
		}
		if wrapped, ok := current.(interface{ Unwrap() error }); ok {
			current = wrapped.Unwrap()
		} else {
			break
		}
	}

	return stack
}

// LoggableError 返回适合日志记录的错误信息
func LoggableError(err error) map[string]interface{} {
	result := make(map[string]interface{})

	if e, ok := err.(*Error); ok {
		result["code"] = e.Code
		result["message"] = e.Message
		result["severity"] = e.Severity
		result["is_temporary"] = e.IsTemporary
		result["timestamp"] = e.Timestamp
		result["caller"] = fmt.Sprintf("%s:%d %s", e.CallerFile, e.CallerLine, e.CallerFunc)

		if e.Err != nil {
			result["cause"] = e.Err.Error()
		}
		if len(e.Context) > 0 {
			result["context"] = e.Context
		}
	} else {
		result["message"] = err.Error()
		result["severity"] = "unknown"
	}

	return result
}

// RecoverAndWrap 从panic中恢复并包装错误
func RecoverAndWrap(code ErrorCode, message string) {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("%v", v)
		}
		panic(Wrap(err, code, message))
	}
}

// SafeExecute 安全执行函数，捕获panic并返回错误
func SafeExecute(fn func() error, code ErrorCode, message string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var panicErr error
			switch v := r.(type) {
			case error:
				panicErr = v
			default:
				panicErr = fmt.Errorf("%v", v)
			}
			err = Wrap(panicErr, code, message)
		}
	}()

	return fn()
}

