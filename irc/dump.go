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
	if codet, ok := err.(ICodeTraverse); ok {
		code = Code(codet.ClosestCode())
		codeStr := codet.GetCodeStr()
		if lenCodeStr := len(codeStr); len(errMsg) > lenCodeStr && errMsg[:lenCodeStr] == codeStr {
			errMsg = errMsg[lenCodeStr:]
		}
	} else if codeg, ok := err.(ICodeGetter); ok {
		code = Code(codeg.GetCode())
		codeStr := codet.GetCodeStr()
		if lenCodeStr := len(codeStr); len(errMsg) > lenCodeStr && errMsg[:lenCodeStr] == codeStr {
			errMsg = errMsg[lenCodeStr:]
		}
	}
	sb.WriteString(errMsg)
	return code, sb.String()
}
