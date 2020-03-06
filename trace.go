package irr

import (
	"path"
	"runtime"
	"strconv"
	"strings"
)

type (
	traceInfo struct {
		FuncName string `json:"func"`
		FileName string `json:"file"`
		Line     int    `json:"line"`
	}
)

func (t traceInfo) String() string {
	sb := strings.Builder{}
	t.writeTo(&sb) // faster than fmt.Sprintf("%s %s:%d", t.FuncName, t.FileName, t.Line) : benchmark 88ms vs 213ms
	return sb.String()
}

func (t traceInfo) writeTo(sb *strings.Builder) {
	sb.WriteString(t.FuncName)
	sb.WriteRune('@')
	sb.WriteString(t.FileName)
	sb.WriteRune(':')
	sb.WriteString(strconv.Itoa(t.Line))
}

func generateStackTrace(skipMore int) *traceInfo {
	pc, _, _, _ := runtime.Caller(1 + skipMore)
	caller := runtime.FuncForPC(pc)
	funcName := caller.Name()
	fileName, line := caller.FileLine(pc)
	return &traceInfo{
		path.Base(funcName),
		fileName,
		line,
	}
}
