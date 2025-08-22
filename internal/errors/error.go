package errors

import (
	"context"
	"fmt"
)

// Frame 堆栈跟踪帧信息
type Frame struct {
	File     string
	Function string
	Line     int
	Package  string // 包名信息
}

// String 返回帧的字符串表示
func (f Frame) String() string {
	return fmt.Sprintf("%s:%d in %s", f.File, f.Line, f.Function)
}

// ErrorImpl 统一的错误实现 - 基于组合模式和适度的职责分离
// 整合了核心数据管理、标签管理等基础功能
type ErrorImpl struct {
	message string
	code    int64
	cause   error
	tags    map[string][]string
	trace   []Frame
	context context.Context
}

// NewErrorImpl 创建新的ErrorImpl实例
func NewErrorImpl(message string, code int64, cause error, tags map[string][]string, trace []Frame, ctx context.Context) *ErrorImpl {
	return &ErrorImpl{
		message: message,
		code:    code,
		cause:   cause,
		tags:    tags,
		trace:   trace,
		context: ctx,
	}
}

// ========== 基础接口实现 ==========

// Error 实现标准error接口
func (e *ErrorImpl) Error() string {
	if e.cause != nil {
		return e.message + ", " + e.cause.Error()
	}
	return e.message
}

// String 基础字符串表示
func (e *ErrorImpl) String() string {
	return e.Error()
}

// Unwrap 实现错误链接口
func (e *ErrorImpl) Unwrap() error {
	return e.cause
}

// ========== 错误码管理 ==========

// Code 获取错误码
func (e *ErrorImpl) Code() int64 {
	return e.code
}

// HasCode 检查是否有错误码
func (e *ErrorImpl) HasCode() bool {
	return e.code != 0
}

// SetCode 设置错误码
func (e *ErrorImpl) SetCode(code int64) *ErrorImpl {
	e.code = code
	return e
}

// ========== 堆栈跟踪管理 ==========

// StackTrace 获取堆栈跟踪
func (e *ErrorImpl) StackTrace() []Frame {
	return e.trace
}

// HasStackTrace 检查是否有堆栈跟踪
func (e *ErrorImpl) HasStackTrace() bool {
	return len(e.trace) > 0
}

// ========== 标签管理 ==========

// Tags 获取所有标签
func (e *ErrorImpl) Tags() map[string][]string {
	if e.tags == nil {
		return make(map[string][]string)
	}
	// 返回副本，避免外部修改
	result := make(map[string][]string)
	for k, v := range e.tags {
		result[k] = append([]string(nil), v...)
	}
	return result
}

// SetTag 设置标签 (void方法)
func (e *ErrorImpl) SetTag(key, value string) {
	if e.tags == nil {
		e.tags = make(map[string][]string)
	}
	e.tags[key] = append(e.tags[key], value)
}

// Tag 设置标签 (链式调用)
func (e *ErrorImpl) Tag(key, value string) *ErrorImpl {
	e.SetTag(key, value)
	return e
}

// GetTag 获取特定标签的值
func (e *ErrorImpl) GetTag(key string) []string {
	if e.tags == nil {
		return nil
	}
	if values, exists := e.tags[key]; exists {
		// 返回副本，避免外部修改
		return append([]string(nil), values...)
	}
	return nil
}

// ========== 上下文管理 ==========

// Context 获取上下文
func (e *ErrorImpl) Context() context.Context {
	if e.context == nil {
		return context.Background()
	}
	return e.context
}

// WithContext 设置上下文
func (e *ErrorImpl) WithContext(ctx context.Context) *ErrorImpl {
	e.context = ctx
	return e
}
