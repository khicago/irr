package irr

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextualError(t *testing.T) {
	ctx := context.Background()

	// 测试基本创建
	err := ErrorWithContext(ctx, "test error")
	assert.NotNil(t, err)
	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, ctx, err.Context())

	// 测试包装
	innerErr := Error("inner error")
	wrappedErr := WrapWithContext(ctx, innerErr, "wrapped")
	assert.Contains(t, wrappedErr.Error(), "wrapped")
	assert.Contains(t, wrappedErr.Error(), "inner error")

	// 测试堆栈跟踪
	tracedErr := TraceWithContext(ctx, "traced error")
	assert.NotNil(t, tracedErr.GetTraceInfo())

	// 测试跟踪包装
	trackedErr := TrackWithContext(ctx, innerErr, "tracked")
	assert.NotNil(t, trackedErr.GetTraceInfo())
	assert.Contains(t, trackedErr.Error(), "tracked")
}

func TestContextualErrorWithValue(t *testing.T) {
	ctx := context.Background()
	err := ErrorWithContext(ctx, "test error")

	// 添加值
	errWithValue := err.WithValue("key", "value")
	extractedCtx := errWithValue.Context()

	assert.Equal(t, "value", extractedCtx.Value("key"))
}

func TestContextualErrorWithTimeout(t *testing.T) {
	ctx := context.Background()
	err := ErrorWithContext(ctx, "test error")

	// 设置超时
	errWithTimeout := err.WithTimeout(time.Second)
	extractedCtx := errWithTimeout.Context()

	deadline, ok := extractedCtx.Deadline()
	assert.True(t, ok)
	assert.True(t, time.Until(deadline) <= time.Second)
}

func TestContextualErrorWithDeadline(t *testing.T) {
	ctx := context.Background()
	err := ErrorWithContext(ctx, "test error")

	deadline := time.Now().Add(time.Hour)
	errWithDeadline := err.WithDeadline(deadline)
	extractedCtx := errWithDeadline.Context()

	ctxDeadline, ok := extractedCtx.Deadline()
	assert.True(t, ok)
	assert.True(t, ctxDeadline.Equal(deadline))
}

func TestContextualErrorToString(t *testing.T) {
	// 创建带截止时间的上下文
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Hour))
	defer cancel()

	err := ErrorWithContext(ctx, "test error")
	toString := err.ToString(false, ", ")

	assert.Contains(t, toString, "test error")
	assert.Contains(t, toString, "deadline:")
}

func TestIsContextError(t *testing.T) {
	ctx := context.Background()

	// 测试 ContextualError
	ctxErr := ErrorWithContext(ctx, "contextual error")
	assert.True(t, IsContextError(ctxErr))

	// 测试普通错误
	normalErr := Error("normal error")
	assert.False(t, IsContextError(normalErr))

	// 测试包装的 ContextualError
	wrappedCtxErr := Wrap(ctxErr, "wrapped")
	assert.True(t, IsContextError(wrappedCtxErr))

	// 测试 nil
	assert.False(t, IsContextError(nil))
}

func TestExtractContext(t *testing.T) {
	originalCtx := context.WithValue(context.Background(), "test", "value")

	// 从 ContextualError 提取
	ctxErr := ErrorWithContext(originalCtx, "error")
	extractedCtx := ExtractContext(ctxErr)
	assert.Equal(t, "value", extractedCtx.Value("test"))

	// 从包装的错误提取
	wrappedErr := Wrap(ctxErr, "wrapped")
	extractedCtx = ExtractContext(wrappedErr)
	assert.Equal(t, "value", extractedCtx.Value("test"))

	// 从普通错误提取（应该返回 Background）
	normalErr := Error("normal")
	extractedCtx = ExtractContext(normalErr)
	assert.Equal(t, context.Background(), extractedCtx)

	// 从 nil 提取
	extractedCtx = ExtractContext(nil)
	assert.Equal(t, context.Background(), extractedCtx)
}
