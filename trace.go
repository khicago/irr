package irr

import (
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type (
	traceInfo struct {
		FuncName string `json:"func"`
		FileName string `json:"file"`
		Line     int    `json:"line"`

		// 缓存字符串表示，避免重复构建
		cached string
		once   sync.Once
	}
)

func (t *traceInfo) String() string {
	t.once.Do(func() {
		var sb strings.Builder
		sb.Grow(len(t.FuncName) + len(t.FileName) + 16) // 预分配容量
		t.writeTo(&sb)
		t.cached = sb.String()
	})
	return t.cached
}

// writeTo a string builder
// faster than fmt.Sprintf("%s %s:%d", t.FuncName, t.FileName, t.Line)
// benchmark 88ms vs 213ms
func (t *traceInfo) writeTo(sb *strings.Builder) *strings.Builder {
	sb.WriteString(t.FuncName)
	sb.WriteRune('@')
	sb.WriteString(t.FileName)
	sb.WriteRune(':')
	sb.WriteString(strconv.Itoa(t.Line))
	return sb
}

var (
	// 堆栈信息缓存池，复用 traceInfo 对象
	tracePool = sync.Pool{
		New: func() interface{} {
			return &traceInfo{}
		},
	}
)

func generateStackTrace(skipMore int) *traceInfo {
	pc, _, _, _ := runtime.Caller(1 + skipMore)
	caller := runtime.FuncForPC(pc)
	funcName := caller.Name()
	fileName, line := caller.FileLine(pc)

	// 从池中获取对象，减少内存分配
	trace := tracePool.Get().(*traceInfo)
	trace.FuncName = path.Base(funcName)
	trace.FileName = fileName
	trace.Line = line
	trace.cached = ""        // 重置缓存
	trace.once = sync.Once{} // 重置once

	return trace
}

// 优化：添加释放方法，虽然在错误处理中不常用，但提供了可能性
func (t *traceInfo) Release() {
	t.FuncName = ""
	t.FileName = ""
	t.Line = 0
	t.cached = ""
	t.once = sync.Once{}
	tracePool.Put(t)
}
