package errors

import (
	"fmt"
	"strings"
)

// formatting.go - 错误格式化和输出相关功能
// 包含所有的错误格式化、详细信息展示等方法

// ========== 格式化接口实现 ==========

// Format 实现fmt.Formatter接口
func (e *ErrorImpl) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, e.Details())
		} else {
			fmt.Fprint(s, e.String())
		}
	case 's':
		fmt.Fprint(s, e.String())
	case 'q':
		fmt.Fprintf(s, "%q", e.String())
	}
}

// ========== 详细信息展示 ==========

// Details 详细信息包含堆栈跟踪
func (e *ErrorImpl) Details() string {
	var parts []string
	parts = append(parts, e.message)

	if e.HasCode() {
		parts = append(parts, fmt.Sprintf("code: %d", e.code))
	}

	if len(e.tags) > 0 {
		var tagParts []string
		for k, values := range e.tags {
			for _, v := range values {
				tagParts = append(tagParts, fmt.Sprintf("%s=%s", k, v))
			}
		}
		parts = append(parts, fmt.Sprintf("tags: [%s]", strings.Join(tagParts, ", ")))
	}

	if e.HasStackTrace() {
		var traceParts []string
		for _, frame := range e.trace {
			traceParts = append(traceParts, frame.String())
		}
		parts = append(parts, fmt.Sprintf("stack:\n%s", strings.Join(traceParts, "\n")))
	}

	return strings.Join(parts, " | ")
}

// StackTraceProvider 定义能提供堆栈跟踪的接口
type StackTraceProvider interface {
	HasStackTrace() bool
	StackTrace() []Frame
}

// MessageProvider 定义能提供消息的接口
type MessageProvider interface {
	Error() string
}

// extractShortFuncName 安全地提取短函数名
func extractShortFuncName(funcName string) string {
	if pathParts := strings.Split(funcName, "/"); len(pathParts) > 0 {
		return pathParts[len(pathParts)-1]
	}
	return funcName
}

// shouldSkipFrame 判断是否应该跳过堆栈帧
func shouldSkipFrame(frame Frame) bool {
	return strings.Contains(frame.Function, "runtime.") ||
		strings.Contains(frame.File, "/runtime/")
}

// formatStackLine 格式化单个堆栈行
func (e *ErrorImpl) formatStackLine(err error, frame Frame) string {
	shortFunc := extractShortFuncName(frame.Function)

	// 根据错误类型选择消息源
	if directErr, ok := err.(*ErrorImpl); ok {
		return directErr.message + " " + shortFunc + " " + frame.String()
	}
	return err.Error() + " " + shortFunc + " " + frame.String()
}

// safeGetStackTraceProvider 安全地获取堆栈跟踪提供者
func safeGetStackTraceProvider(err error) (StackTraceProvider, bool) {
	switch v := err.(type) {
	case *ErrorImpl:
		return v, v.HasStackTrace()
	case StackTraceProvider:
		return v, v.HasStackTrace()
	default:
		return nil, false
	}
}

// collectStackLines 收集错误链中的堆栈行
func (e *ErrorImpl) collectStackLines(split string) []string {
	var stackLines []string
	current := error(e)

	for current != nil {
		provider, hasTrace := safeGetStackTraceProvider(current)
		if !hasTrace {
			current = e.unwrapSafely(current)
			continue
		}

		// 处理堆栈跟踪
		traces := provider.StackTrace()
		for _, frame := range traces {
			if shouldSkipFrame(frame) {
				continue
			}

			stackLine := e.formatStackLine(current, frame)
			stackLines = append(stackLines, stackLine)
			break // 只取第一个有效用户帧
		}

		current = e.unwrapSafely(current)
	}

	return stackLines
}

// unwrapSafely 安全地解包错误
func (e *ErrorImpl) unwrapSafely(err error) error {
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		return unwrapper.Unwrap()
	}
	return nil
}

// ToString 兼容方法 - 格式化字符串输出
func (e *ErrorImpl) ToString(printTrace bool, split string) string {
	if printTrace && e.HasStackTrace() {
		stackLines := e.collectStackLines(split)
		if len(stackLines) > 0 {
			return strings.Join(stackLines, split)
		}
	}

	// 如果有原因错误，使用自定义分隔符
	if e.cause != nil {
		return e.message + split + e.cause.Error()
	}
	return e.message
}
