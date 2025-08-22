package metrics

import (
	"testing"
	"time"

	"github.com/khicago/irr"
)

func TestErrorAnalyzer_Basic(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	// 测试分析标准错误
	err := irr.New("test error").Code(404).Tag("module", "test").Build()
	analyzer.Analyze(err)

	stats := analyzer.GetStats()

	// 检查基础统计
	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors=1, got %d", stats.TotalErrors)
	}

	if stats.ErrorsWithCode != 1 {
		t.Errorf("Expected ErrorsWithCode=1, got %d", stats.ErrorsWithCode)
	}

	// 检查错误码分布
	if stats.CodeDistribution[404] != 1 {
		t.Errorf("Expected code 404 count=1, got %d", stats.CodeDistribution[404])
	}

	// 检查错误类型分布
	if stats.TypeDistribution["irr.IRR"] != 1 {
		t.Errorf("Expected irr.IRR count=1, got %d", stats.TypeDistribution["irr.IRR"])
	}
}

func TestErrorAnalyzer_WithTrace(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	err := irr.New("trace error").Trace().Build()
	analyzer.Analyze(err)

	stats := analyzer.GetStats()

	if stats.ErrorsWithTrace != 1 {
		t.Errorf("Expected ErrorsWithTrace=1, got %d", stats.ErrorsWithTrace)
	}
}

func TestErrorAnalyzer_WithWrapped(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	originalErr := irr.New("original").Build()
	wrappedErr := irr.Wraps(originalErr, "wrapped").Build()

	analyzer.Analyze(wrappedErr)

	stats := analyzer.GetStats()

	if stats.WrappedErrors != 1 {
		t.Errorf("Expected WrappedErrors=1, got %d", stats.WrappedErrors)
	}
}

func TestErrorAnalyzer_TagStats(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	err := irr.New("tagged error").
		Tag("module", "auth").
		Tag("severity", "high").
		Tag("module", "user"). // 同一个key的多个值
		Build()

	// 检查错误是否实现了正确的接口
	if taggable, ok := err.(irr.IRR); ok {
		tags := taggable.Tags()
		t.Logf("Error tags: %+v", tags)

		if len(tags) == 0 {
			t.Error("Error has no tags")
		}
	} else {
		t.Error("Error does not implement TaggableError")
	}

	analyzer.Analyze(err)

	stats := analyzer.GetStats()

	t.Logf("TagStats: %+v", stats.TagStats)

	// 检查标签统计
	if len(stats.TagStats) != 2 { // module 和 severity
		t.Errorf("Expected 2 tag keys, got %d, TagStats: %+v", len(stats.TagStats), stats.TagStats)
	}

	// 检查 module 标签的值
	moduleStats := stats.TagStats["module"]
	if moduleStats == nil {
		t.Error("module tag stats is nil")
	} else {
		if moduleStats["auth"] != 1 {
			t.Errorf("Expected module=auth count=1, got %d", moduleStats["auth"])
		}
		if moduleStats["user"] != 1 {
			t.Errorf("Expected module=user count=1, got %d", moduleStats["user"])
		}
	}

	// 检查 severity 标签
	severityStats := stats.TagStats["severity"]
	if severityStats == nil {
		t.Error("severity tag stats is nil")
	} else {
		if severityStats["high"] != 1 {
			t.Errorf("Expected severity=high count=1, got %d", severityStats["high"])
		}
	}
}

func TestErrorAnalyzer_Reset(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	// 添加一些统计
	err := irr.New("test").Code(500).Build()
	analyzer.Analyze(err)

	// 验证有统计数据
	stats := analyzer.GetStats()
	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors=1 before reset, got %d", stats.TotalErrors)
	}

	// 重置
	analyzer.Reset()

	// 验证统计已重置
	stats = analyzer.GetStats()
	if stats.TotalErrors != 0 {
		t.Errorf("Expected TotalErrors=0 after reset, got %d", stats.TotalErrors)
	}

	if len(stats.CodeDistribution) != 0 {
		t.Errorf("Expected empty CodeDistribution after reset, got %d entries", len(stats.CodeDistribution))
	}
}

func TestErrorAnalyzer_Summary(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	// 添加多种错误
	errors := []error{
		irr.New("error1").Code(404).Build(),
		irr.New("error2").Code(500).Build(),
		irr.New("error3").Code(404).Build(), // 重复的错误码
		irr.New("error4").Build(),           // 无错误码
	}

	for _, err := range errors {
		analyzer.Analyze(err)
		time.Sleep(1 * time.Millisecond) // 确保时间差
	}

	summary := analyzer.Summary()

	if summary.TotalErrors != 4 {
		t.Errorf("Expected TotalErrors=4, got %d", summary.TotalErrors)
	}

	if summary.Duration <= 0 {
		t.Errorf("Expected positive duration, got %v", summary.Duration)
	}

	if summary.ErrorRate <= 0 {
		t.Errorf("Expected positive error rate, got %f", summary.ErrorRate)
	}

	// 检查 Top Error Codes
	if len(summary.TopErrorCodes) == 0 {
		t.Error("Expected TopErrorCodes to contain entries")
	}

	// 404 应该是最频繁的错误码（出现2次）
	if summary.TopErrorCodes[0].Code != 404 || summary.TopErrorCodes[0].Count != 2 {
		t.Errorf("Expected top error code 404 with count 2, got code %d with count %d",
			summary.TopErrorCodes[0].Code, summary.TopErrorCodes[0].Count)
	}
}

func TestErrorAnalyzer_MaxLimits(t *testing.T) {
	analyzer := NewErrorAnalyzer().WithMaxCodeStats(2).WithMaxTagStats(2)

	// 添加超过限制的错误码
	for i := int64(1); i <= 5; i++ {
		err := irr.New("error").Code(i).Build()
		analyzer.Analyze(err)
	}

	stats := analyzer.GetStats()

	// 应该只记录前两种错误码
	if len(stats.CodeDistribution) > 2 {
		t.Errorf("Expected max 2 code entries due to limit, got %d", len(stats.CodeDistribution))
	}

	// 测试标签限制
	analyzer.Reset()
	err := irr.New("tagged error")
	for i := 1; i <= 5; i++ {
		err = err.Tag("test", "value"+string(rune('0'+i)))
	}
	builtErr := err.Build()
	analyzer.Analyze(builtErr)

	stats = analyzer.GetStats()
	testTagStats := stats.TagStats["test"]
	if len(testTagStats) > 2 {
		t.Errorf("Expected max 2 tag values due to limit, got %d", len(testTagStats))
	}
}

func TestErrorAnalyzer_StandardError(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	// 测试标准错误
	err := &customError{msg: "custom error"}
	analyzer.Analyze(err)

	stats := analyzer.GetStats()

	if stats.TotalErrors != 1 {
		t.Errorf("Expected TotalErrors=1, got %d", stats.TotalErrors)
	}

	if stats.TypeDistribution["standard_error"] != 1 {
		t.Errorf("Expected standard_error count=1, got %d", stats.TypeDistribution["standard_error"])
	}
}

// 辅助类型用于测试
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func TestErrorAnalyzer_ConcurrentAccess(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	// 并发添加错误
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				err := irr.New("concurrent error").Code(int64(id)).Build()
				analyzer.Analyze(err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := analyzer.GetStats()

	if stats.TotalErrors != 1000 {
		t.Errorf("Expected TotalErrors=1000, got %d", stats.TotalErrors)
	}
}
