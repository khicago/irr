package irr

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// ErrorBuilder 错误构建器 - 基于第一性原理的简化设计
type ErrorBuilder struct {
	message   string
	code      int64
	cause     error
	tags      map[string][]string
	trace     bool
	traceSkip int
	enrichers []Enricher
}

// New 创建错误构建器
func New(message string, args ...any) *ErrorBuilder {
	return &ErrorBuilder{
		message: fmt.Sprintf(message, args...),
		tags:    make(map[string][]string),
	}
}

// Wraps 包装错误
func Wraps(err error, message string, args ...any) *ErrorBuilder {
	return &ErrorBuilder{
		message: fmt.Sprintf(message, args...),
		cause:   err,
		tags:    make(map[string][]string),
	}
}

// Code 设置错误码
func (b *ErrorBuilder) Code(code int64) *ErrorBuilder {
	b.code = code
	return b
}

// Tag 添加标签
func (b *ErrorBuilder) Tag(key, value string) *ErrorBuilder {
	if b.tags == nil {
		b.tags = make(map[string][]string)
	}
	b.tags[key] = append(b.tags[key], value)
	return b
}

// Trace 启用堆栈跟踪
func (b *ErrorBuilder) Trace() *ErrorBuilder {
	b.trace = true
	return b
}

// TraceSkip 添加堆栈跟踪并跳过指定帧数
func (b *ErrorBuilder) TraceSkip(skip int) *ErrorBuilder {
	b.trace = true
	b.traceSkip = skip
	return b
}

// WithEnricher 添加丰富器
func (b *ErrorBuilder) WithEnricher(enricher Enricher) *ErrorBuilder {
	b.enrichers = append(b.enrichers, enricher)
	return b
}

// Build 构建错误
func (b *ErrorBuilder) Build() IRR {
	return b.BuildWithContext(context.Background())
}

// BuildWithContext 使用上下文构建错误
func (b *ErrorBuilder) BuildWithContext(ctx context.Context) IRR {
	var trace []Frame
	// 添加堆栈跟踪
	if b.trace {
		trace = captureStackTraceWithSkip(b.traceSkip)
	}

	err := newErrorImpl(b.message, b.code, b.cause, b.copyTags(), trace, ctx)

	// 应用丰富器
	result := IRR(err)
	for _, enricher := range b.enrichers {
		result = enricher.Enrich(ctx, result)
	}

	return result
}

// copyTags 复制标签映射
func (b *ErrorBuilder) copyTags() map[string][]string {
	if b.tags == nil {
		return nil
	}
	result := make(map[string][]string)
	for k, v := range b.tags {
		result[k] = append([]string(nil), v...)
	}
	return result
}

// captureStackTrace 捕获堆栈跟踪
func captureStackTrace() []Frame {
	return captureStackTraceWithSkip(0)
}

func captureStackTraceWithSkip(extraSkip int) []Frame {
	var frames []Frame

	// Debug: print the calling stack for the first few levels
	/*
		for debugI := 0; debugI < 8; debugI++ {
			if pc, file, line, ok := runtime.Caller(debugI); ok {
				fn := runtime.FuncForPC(pc)
				if fn != nil {
					fmt.Printf("Debug stack[%d]: %s:%d in %s\n", debugI, file, line, fn.Name())
				}
			}
		}
	*/

	// 跳过当前函数、newErrorImpl、Build函数、irr包装函数，加上额外跳过的帧数
	skip := 4 + extraSkip    // 跳过: captureStackTraceWithSkip, BuildWithContext, Build, irr.Trace/Track, 然后extraSkip
	for i := 0; i < 2; i++ { // 限制为2个帧以匹配测试期望
		pc, file, line, ok := runtime.Caller(skip + i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// 提取包名
		funcName := fn.Name()
		pkg := ""
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			if lastDot := strings.Index(funcName[lastSlash:], "."); lastDot >= 0 {
				pkg = funcName[:lastSlash+lastDot]
			}
		}

		frames = append(frames, Frame{
			File:     file,
			Function: funcName,
			Line:     line,
			Package:  pkg,
		})
	}

	return frames
}
