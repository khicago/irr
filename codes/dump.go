package codes

import (
	"fmt"
	"github.com/khicago/irr"
	"strings"
)

// findNearestCode 在错误链中查找最近的错误码
func findNearestCode(err error) int64 {
	current := err
	for current != nil {
		if irrErr, ok := current.(irr.IRR); ok && irrErr.HasCode() {
			return irrErr.Code()
		}

		// 尝试解包错误
		if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			current = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return 0
}

// DumpToCodeNError 从错误中提取错误码和消息
// 基于第一性原理重新设计，使用统一的IRR接口
func DumpToCodeNError(succ, unknown Code, err error, msgOrFmt string, args ...any) (code Code, msg string) {
	if err == nil {
		return succ, ""
	}

	sb := strings.Builder{}
	if msgOrFmt != "" {
		if len(args) > 0 {
			sb.WriteString(fmt.Sprintf(msgOrFmt, args...))
		} else {
			sb.WriteString(msgOrFmt)
		}
		sb.WriteString(", ")
	}

	code = unknown
	errMsg := err.Error()

	// 使用统一的IRR接口获取错误码
	if irrErr, ok := err.(irr.IRR); ok {
		// 尝试获取最近的错误码，包括错误链中的错误码
		if nearestCode := findNearestCode(err); nearestCode != 0 {
			code = Code(nearestCode)
			// 简化错误消息处理 - 直接使用错误消息，不做复杂的字符串处理
			errMsg = irrErr.Error()
		}
	}

	sb.WriteString(errMsg)
	return code, sb.String()
}

// ExtractCode 从错误中提取错误码，如果没有则返回默认值
func ExtractCode(err error, defaultCode Code) Code {
	if err == nil {
		return defaultCode
	}

	if irrErr, ok := err.(irr.IRR); ok && irrErr.HasCode() {
		return Code(irrErr.Code())
	}

	return defaultCode
}

// HasErrorCode 检查错误是否包含特定错误码
func HasErrorCode(err error, code Code) bool {
	if err == nil {
		return false
	}

	if irrErr, ok := err.(irr.IRR); ok {
		return irrErr.HasCode() && irrErr.Code() == int64(code)
	}

	return false
}
