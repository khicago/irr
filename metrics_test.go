package irr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	// 重置统计
	ResetMetrics()

	// 创建一些错误来测试统计
	_ = Error("test error")              // 1 error created
	_ = ErrorC(1001, "error with code")  // 1 error created, 1 with code
	_ = Trace("trace error")             // 1 error created, 1 with trace
	_ = Wrap(Error("inner"), "wrapped")  // 2 errors created (inner + wrap), 1 wrapped
	_ = Track(Error("inner"), "tracked") // 2 errors created (inner + track), 1 wrapped, 1 with trace

	// 执行一些遍历操作
	err := Track(Error("inner"), "outer") // 2 more errors created, 1 more wrapped, 1 more with trace
	_ = err.TraverseToRoot(func(e error) error { return nil })
	_ = err.TraverseToSource(func(e error, isSource bool) error { return nil })

	// 获取统计信息
	metrics := GetMetrics()

	// 验证统计 - 总共应该有 8 个错误创建
	// Error: 1, ErrorC: 1, Trace: 1, Wrap(Error): 2, Track(Error): 2, Track(Error) again: 2 = 9
	assert.Equal(t, int64(9), metrics.ErrorCreated, "错误创建数量不匹配")
	assert.Equal(t, int64(1), metrics.ErrorWithCode, "带错误码的错误数量不匹配")
	assert.Equal(t, int64(3), metrics.ErrorWithTrace, "带堆栈跟踪的错误数量不匹配")
	assert.Equal(t, int64(3), metrics.ErrorWrapped, "包装错误数量不匹配")
	assert.Equal(t, int64(2), metrics.TraverseOps, "遍历操作数量不匹配")

	// 验证错误码统计
	assert.Equal(t, int64(1), metrics.CodeStats[1001], "错误码1001的统计不匹配")

	// 验证时间戳
	assert.True(t, time.Since(metrics.LastErrorTime) < time.Second, "最后错误时间不正确")
}

func TestMetricsReset(t *testing.T) {
	// 创建一些错误
	_ = Error("test")
	_ = ErrorC(2001, "test")

	// 验证有统计
	metrics := GetMetrics()
	assert.Greater(t, metrics.ErrorCreated, int64(0))
	assert.Greater(t, len(metrics.CodeStats), 0)

	// 重置统计
	ResetMetrics()

	// 验证已重置
	metrics = GetMetrics()
	assert.Equal(t, int64(0), metrics.ErrorCreated)
	assert.Equal(t, 0, len(metrics.CodeStats))
	assert.True(t, metrics.LastErrorTime.IsZero())
}

func TestMetricsConcurrent(t *testing.T) {
	ResetMetrics()

	// 并发创建错误
	done := make(chan bool)
	for i := 1; i <= 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				_ = ErrorC(int64(id), "error %d-%d", id, j)
			}
		}(i)
	}

	// 等待所有协程完成
	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := GetMetrics()
	assert.Equal(t, int64(1000), metrics.ErrorCreated)
	assert.Equal(t, int64(1000), metrics.ErrorWithCode)
	assert.Equal(t, 10, len(metrics.CodeStats))

	// 验证每个错误码的统计
	for i := 1; i <= 10; i++ {
		assert.Equal(t, int64(100), metrics.CodeStats[int64(i)], "错误码 %d 的统计不正确", i)
	}
}
