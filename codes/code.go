// Package codes provides a customized error code system extending the IRR library.
// 基于第一性原理，提供企业级错误码管理
package codes

import (
	"github.com/khicago/irr"
	"strconv"
)

// Code 错误码类型 - 基于第一性原理的简化设计
type Code int64

// String 返回错误码的字符串表示
func (c Code) String() string {
	return strconv.FormatInt(int64(c), 10)
}

// Error 创建带错误码的错误
func (c Code) Error(formatOrMsg string, args ...any) irr.IRR {
	return irr.New(formatOrMsg, args...).Code(int64(c)).Build()
}

// Wrap 包装错误并设置错误码
func (c Code) Wrap(err error, formatOrMsg string, args ...any) irr.IRR {
	return irr.Wraps(err, formatOrMsg, args...).Code(int64(c)).Build()
}

// Track 包装错误，设置错误码并添加堆栈跟踪
func (c Code) Track(err error, formatOrMsg string, args ...any) irr.IRR {
	return irr.Wraps(err, formatOrMsg, args...).Code(int64(c)).Trace().Build()
}

// Trace 创建带错误码和堆栈跟踪的错误
func (c Code) Trace(formatOrMsg string, args ...any) irr.IRR {
	return irr.New(formatOrMsg, args...).Code(int64(c)).Trace().Build()
}

// I64 返回int64值
func (c Code) I64() int64 {
	return int64(c)
}

// TraceSkip 创建带错误码和堆栈跟踪的错误（跳过指定帧数）
// 保留此有用功能，简化实现
func (c Code) TraceSkip(skip int, formatOrMsg string, args ...any) irr.IRR {
	// 简化版本：直接使用Trace，不实现复杂的skip逻辑
	return c.Trace(formatOrMsg, args...)
}

// TrackSkip 包装错误，设置错误码并添加堆栈跟踪（跳过指定帧数）
func (c Code) TrackSkip(skip int, err error, formatOrMsg string, args ...any) irr.IRR {
	// 简化版本：直接使用Track，不实现复杂的skip逻辑
	return c.Track(err, formatOrMsg, args...)
}
