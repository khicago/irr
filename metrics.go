package irr

import (
	"sync"
	"sync/atomic"
	"time"
)

// ErrorMetrics 错误统计信息
type ErrorMetrics struct {
	// 错误创建统计
	ErrorCreated   int64 `json:"error_created"`
	ErrorWithCode  int64 `json:"error_with_code"`
	ErrorWithTrace int64 `json:"error_with_trace"`
	ErrorWrapped   int64 `json:"error_wrapped"`

	// 错误遍历统计
	TraverseOps int64 `json:"traverse_ops"`

	// 时间统计
	LastErrorTime time.Time `json:"last_error_time"`

	// 错误码统计
	CodeStats      map[int64]int64 `json:"code_stats"`
	codeStatsMutex sync.RWMutex    `json:"-"`
}

var globalMetrics = &ErrorMetrics{
	CodeStats: make(map[int64]int64),
}

// GetMetrics 获取全局错误统计信息
func GetMetrics() *ErrorMetrics {
	globalMetrics.codeStatsMutex.RLock()
	defer globalMetrics.codeStatsMutex.RUnlock()

	// 返回副本以避免并发问题
	result := &ErrorMetrics{
		ErrorCreated:   atomic.LoadInt64(&globalMetrics.ErrorCreated),
		ErrorWithCode:  atomic.LoadInt64(&globalMetrics.ErrorWithCode),
		ErrorWithTrace: atomic.LoadInt64(&globalMetrics.ErrorWithTrace),
		ErrorWrapped:   atomic.LoadInt64(&globalMetrics.ErrorWrapped),
		TraverseOps:    atomic.LoadInt64(&globalMetrics.TraverseOps),
		LastErrorTime:  globalMetrics.LastErrorTime,
		CodeStats:      make(map[int64]int64, len(globalMetrics.CodeStats)),
	}

	for code, count := range globalMetrics.CodeStats {
		result.CodeStats[code] = count
	}

	return result
}

// ResetMetrics 重置统计信息
func ResetMetrics() {
	atomic.StoreInt64(&globalMetrics.ErrorCreated, 0)
	atomic.StoreInt64(&globalMetrics.ErrorWithCode, 0)
	atomic.StoreInt64(&globalMetrics.ErrorWithTrace, 0)
	atomic.StoreInt64(&globalMetrics.ErrorWrapped, 0)
	atomic.StoreInt64(&globalMetrics.TraverseOps, 0)

	globalMetrics.codeStatsMutex.Lock()
	globalMetrics.CodeStats = make(map[int64]int64)
	globalMetrics.LastErrorTime = time.Time{}
	globalMetrics.codeStatsMutex.Unlock()
}

// 内部统计函数
func recordErrorCreated() {
	atomic.AddInt64(&globalMetrics.ErrorCreated, 1)
	globalMetrics.LastErrorTime = time.Now()
}

func recordErrorWithCode(code int64) {
	atomic.AddInt64(&globalMetrics.ErrorWithCode, 1)

	globalMetrics.codeStatsMutex.Lock()
	globalMetrics.CodeStats[code]++
	globalMetrics.codeStatsMutex.Unlock()
}

func recordErrorWithTrace() {
	atomic.AddInt64(&globalMetrics.ErrorWithTrace, 1)
}

func recordErrorWrapped() {
	atomic.AddInt64(&globalMetrics.ErrorWrapped, 1)
}

func recordTraverseOp() {
	atomic.AddInt64(&globalMetrics.TraverseOps, 1)
}

// ErrorStatsLogger 错误统计日志接口
type ErrorStatsLogger interface {
	LogErrorStats(metrics *ErrorMetrics)
}

// LogStats 记录错误统计信息到日志
func LogStats(logger ErrorStatsLogger) {
	if logger != nil {
		logger.LogErrorStats(GetMetrics())
	}
}
