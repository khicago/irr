package irc

import (
	"fmt"
	"strings"
)

func DumpToCodeNError(succ, unknown Code, err error, msgOrFmt string, args ...any) (code Code, msg string) {
	if err == nil {
		return succ, ""
	}

	sb := strings.Builder{}
	appendMsg := msgOrFmt
	if appendMsg != "" && len(args) > 0 {
		appendMsg = fmt.Sprintf(msgOrFmt, args...)
		sb.WriteString(appendMsg)
		sb.WriteString(", ")
	}

	code = unknown
	errMsg := err.Error()

	// 优先使用新的NearestCode API
	if codet, ok := err.(interface{ NearestCode() int64 }); ok {
		code = Code(codet.NearestCode())
		if codeStr, ok := err.(interface{ GetCodeStr() string }); ok {
			if lenCodeStr := len(codeStr.GetCodeStr()); len(errMsg) > lenCodeStr && errMsg[:lenCodeStr] == codeStr.GetCodeStr() {
				errMsg = errMsg[lenCodeStr:]
			}
		}
	} else if codet, ok := err.(ICodeTraverse); ok {
		// 向后兼容：使用ClosestCode
		code = Code(codet.ClosestCode())
		codeStr := codet.GetCodeStr()
		if lenCodeStr := len(codeStr); len(errMsg) > lenCodeStr && errMsg[:lenCodeStr] == codeStr {
			errMsg = errMsg[lenCodeStr:]
		}
	} else if codeg, ok := err.(ICodeGetter); ok {
		// 向后兼容：使用GetCode
		code = Code(codeg.GetCode())
		codeStr := codeg.GetCodeStr()
		if lenCodeStr := len(codeStr); len(errMsg) > lenCodeStr && errMsg[:lenCodeStr] == codeStr {
			errMsg = errMsg[lenCodeStr:]
		}
	}
	sb.WriteString(errMsg)
	return code, sb.String()
}
