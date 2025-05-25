package irr

import (
	"context"
	"errors"
	"time"
)

// ContextualError 带上下文的错误接口
type ContextualError interface {
	IRR
	Context() context.Context
	WithContext(ctx context.Context) ContextualError
	WithDeadline(deadline time.Time) ContextualError
	WithTimeout(timeout time.Duration) ContextualError
	WithValue(key, val interface{}) ContextualError
}

// ContextualIrr 实现了带上下文的错误
type ContextualIrr struct {
	*BasicIrr
	ctx context.Context
}

// 确保实现接口
var _ ContextualError = (*ContextualIrr)(nil)

// ErrorWithContext 创建带上下文的错误
func ErrorWithContext(ctx context.Context, formatOrMsg string, args ...any) ContextualError {
	recordErrorCreated()
	err := newBasicIrr(formatOrMsg, args...)
	return &ContextualIrr{
		BasicIrr: err,
		ctx:      ctx,
	}
}

// WrapWithContext 包装错误并添加上下文
func WrapWithContext(ctx context.Context, innerErr error, formatOrMsg string, args ...any) ContextualError {
	recordErrorCreated()
	recordErrorWrapped()
	err := newBasicIrr(formatOrMsg, args...)
	err.inner = innerErr
	return &ContextualIrr{
		BasicIrr: err,
		ctx:      ctx,
	}
}

// TraceWithContext 创建带堆栈跟踪和上下文的错误
func TraceWithContext(ctx context.Context, formatOrMsg string, args ...any) ContextualError {
	recordErrorCreated()
	recordErrorWithTrace()
	err := newBasicIrr(formatOrMsg, args...)
	err.Trace = createTraceInfo(1, nil)
	return &ContextualIrr{
		BasicIrr: err,
		ctx:      ctx,
	}
}

// TrackWithContext 包装错误并添加堆栈跟踪和上下文
func TrackWithContext(ctx context.Context, innerErr error, formatOrMsg string, args ...any) ContextualError {
	recordErrorCreated()
	recordErrorWrapped()
	recordErrorWithTrace()
	err := newBasicIrr(formatOrMsg, args...)
	err.inner = innerErr
	err.Trace = createTraceInfo(1, innerErr)
	return &ContextualIrr{
		BasicIrr: err,
		ctx:      ctx,
	}
}

// Context 返回关联的上下文
func (ce *ContextualIrr) Context() context.Context {
	if ce.ctx == nil {
		return context.Background()
	}
	return ce.ctx
}

// WithContext 使用新的上下文创建副本
func (ce *ContextualIrr) WithContext(ctx context.Context) ContextualError {
	return &ContextualIrr{
		BasicIrr: ce.BasicIrr,
		ctx:      ctx,
	}
}

// WithDeadline 设置截止时间
func (ce *ContextualIrr) WithDeadline(deadline time.Time) ContextualError {
	ctx, _ := context.WithDeadline(ce.Context(), deadline)
	return ce.WithContext(ctx)
}

// WithTimeout 设置超时时间
func (ce *ContextualIrr) WithTimeout(timeout time.Duration) ContextualError {
	ctx, _ := context.WithTimeout(ce.Context(), timeout)
	return ce.WithContext(ctx)
}

// WithValue 添加键值对
func (ce *ContextualIrr) WithValue(key, val interface{}) ContextualError {
	return ce.WithContext(context.WithValue(ce.Context(), key, val))
}

// ToString 重写以包含上下文信息
func (ce *ContextualIrr) ToString(printTrace bool, split string) string {
	result := ce.BasicIrr.ToString(printTrace, split)

	// 添加上下文信息
	if ce.ctx != nil {
		if deadline, ok := ce.ctx.Deadline(); ok {
			result += " [deadline:" + deadline.Format(time.RFC3339) + "]"
		}
		if ce.ctx.Err() != nil {
			result += " [ctx-err:" + ce.ctx.Err().Error() + "]"
		}
	}

	return result
}

// IsContextError 检查是否为上下文相关错误
func IsContextError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(ContextualError); ok {
		return true
	}
	// 递归检查包装的错误
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return IsContextError(unwrapped)
	}
	return false
}

// ExtractContext 从错误中提取上下文
func ExtractContext(err error) context.Context {
	if ce, ok := err.(ContextualError); ok {
		return ce.Context()
	}
	// 递归检查包装的错误
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return ExtractContext(unwrapped)
	}
	return context.Background()
}
