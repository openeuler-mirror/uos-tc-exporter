// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package errors

import (
	"fmt"
	"strings"
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
	ErrCodeBusiness    ErrorCode = 4000
	ErrCodeMetrics     ErrorCode = 4001
	ErrCodeLandingPage ErrorCode = 4002
)

// Error 自定义错误结构
type Error struct {
	Code    ErrorCode
	Message string
	Err     error
	Context map[string]interface{}
}

// New 创建新的错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap 包装现有错误
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}

	var customErr *Error
	if e, ok := err.(*Error); ok {
		customErr = e
	} else {
		customErr = &Error{
			Code:    code,
			Message: message,
			Err:     err,
			Context: make(map[string]interface{}),
		}
	}

	return customErr
}

// WithContext 添加错误上下文
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
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
	var parts []string

	if e.Code != 0 {
		parts = append(parts, fmt.Sprintf("[%d]", e.Code))
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Err != nil {
		parts = append(parts, fmt.Sprintf("caused by: %v", e.Err))
	}

	if len(e.Context) > 0 {
		contextStr := fmt.Sprintf("context: %v", e.Context)
		parts = append(parts, contextStr)
	}

	return strings.Join(parts, " ")
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
func (e *Error) GetContext() map[string]interface{} {
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
func GetErrorContext(err error) map[string]interface{} {
	if e, ok := err.(*Error); ok {
		return e.Context
	}
	return nil
}
