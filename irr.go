package irr

// 基于第一性原理的便利函数
// 提供最常用的错误创建方式，内部使用构建器

// Error 创建简单错误
func Error(formatOrMsg string, args ...any) IRR {
	return New(formatOrMsg, args...).Build()
}

// ErrorC 创建带错误码的错误
func ErrorC[T int64](code T, formatOrMsg string, args ...any) IRR {
	return New(formatOrMsg, args...).Code(int64(code)).Build()
}

// Wrap 包装错误
func Wrap(err error, formatOrMsg string, args ...any) IRR {
	return Wraps(err, formatOrMsg, args...).Build()
}

// Track 包装错误并添加堆栈跟踪
func Track(err error, formatOrMsg string, args ...any) IRR {
	return Wraps(err, formatOrMsg, args...).Trace().Build()
}

// Trace 创建带堆栈跟踪的错误
func Trace(formatOrMsg string, args ...any) IRR {
	return New(formatOrMsg, args...).Trace().Build()
}

// TraceSkip 创建带堆栈跟踪的错误（跳过指定帧数）
// 保留此有用功能，简化实现
func TraceSkip(skip int, formatOrMsg string, args ...any) IRR {
	return New(formatOrMsg, args...).TraceSkip(skip).Build()
}

// TrackSkip 包装错误并添加堆栈跟踪（跳过指定帧数）
func TrackSkip(skip int, err error, formatOrMsg string, args ...any) IRR {
	return Wraps(err, formatOrMsg, args...).TraceSkip(skip).Build()
}
