// Package metrics 提供错误分析和监控的辅助工具
// 基于第一性原理简化并发机制
package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/khicago/irr"
)

// ErrorStats 错误统计信息
type ErrorStats struct {
	// 基础统计
	TotalErrors     int64 `json:"total_errors"`
	ErrorsWithCode  int64 `json:"errors_with_code"`
	ErrorsWithTrace int64 `json:"errors_with_trace"`
	WrappedErrors   int64 `json:"wrapped_errors"`

	// 时间统计
	LastErrorTime  time.Time `json:"last_error_time"`
	FirstErrorTime time.Time `json:"first_error_time"`

	// 错误码分布
	CodeDistribution map[int64]int64 `json:"code_distribution"`

	// 错误类型分布
	TypeDistribution map[string]int64 `json:"type_distribution"`

	// 标签统计
	TagStats map[string]map[string]int64 `json:"tag_stats"`
}

// AnalysisSummary 分析摘要 - 提供有用的统计信息
type AnalysisSummary struct {
	TotalErrors        int64
	ErrorRate          float64          // 错误率（基于时间窗口）
	Duration           time.Duration    // 统计时间跨度
	TopErrorCodes      []CodeStat       // 最常见的错误码
	MostFrequentErrors map[string]int64 // 最频繁的错误类型
}

// CodeStat 错误码统计
type CodeStat struct {
	Code  int64 `json:"code"`
	Count int64 `json:"count"`
}

// ErrorAnalyzer 错误分析器 - 基于第一性原理简化并发
// 统一使用mutex，避免atomic/mutex混合的复杂性
type ErrorAnalyzer struct {
	stats *ErrorStats
	mu    sync.RWMutex // 统一的并发保护

	// 配置项
	maxCodeStats int // 最多记录多少种错误码
	maxTagStats  int // 最多记录多少种标签值
}

// NewErrorAnalyzer 创建错误分析器
func NewErrorAnalyzer() *ErrorAnalyzer {
	now := time.Now()
	return &ErrorAnalyzer{
		stats: &ErrorStats{
			FirstErrorTime:   now,
			LastErrorTime:    now,
			CodeDistribution: make(map[int64]int64),
			TypeDistribution: make(map[string]int64),
			TagStats:         make(map[string]map[string]int64),
		},
		maxCodeStats: 1000, // 默认最多记录1000种错误码
		maxTagStats:  100,  // 默认每个标签最多记录100种值
	}
}

// WithMaxCodeStats 设置最大错误码统计数
func (ea *ErrorAnalyzer) WithMaxCodeStats(max int) *ErrorAnalyzer {
	ea.maxCodeStats = max
	return ea
}

// WithMaxTagStats 设置最大标签统计数
func (ea *ErrorAnalyzer) WithMaxTagStats(max int) *ErrorAnalyzer {
	ea.maxTagStats = max
	return ea
}

// Analyze 分析错误并更新统计 - 简化并发设计
func (ea *ErrorAnalyzer) Analyze(err error) {
	if err == nil {
		return
	}

	ea.mu.Lock()
	defer ea.mu.Unlock()

	// 统一使用mutex保护，简化并发模型
	ea.stats.TotalErrors++
	ea.stats.LastErrorTime = time.Now()

	// 分析 IRR 错误
	if irrErr, ok := err.(irr.IRR); ok {
		ea.analyzeIRRError(irrErr)
	} else {
		// 处理标准错误
		ea.analyzeStandardError(err)
	}
}

// analyzeIRRError 分析 IRR 错误（调用时已持有锁）
func (ea *ErrorAnalyzer) analyzeIRRError(err irr.IRR) {
	// 检查是否有错误码
	if err.HasCode() {
		ea.stats.ErrorsWithCode++
		code := err.Code()

		// 限制错误码统计数量，避免内存泄露
		if len(ea.stats.CodeDistribution) < ea.maxCodeStats {
			ea.stats.CodeDistribution[code]++
		}
	}

	// 检查是否有堆栈跟踪
	if err.HasStackTrace() {
		ea.stats.ErrorsWithTrace++
	}

	// 检查是否是包装错误
	if err.Unwrap() != nil {
		ea.stats.WrappedErrors++
	}

	// 记录错误类型 - 使用接口名而非具体类型
	ea.stats.TypeDistribution["irr.IRR"]++

	// 分析标签
	ea.analyzeTags(err.Tags())
}

// analyzeStandardError 分析标准错误（调用时已持有锁）
func (ea *ErrorAnalyzer) analyzeStandardError(err error) {
	// 记录错误类型 - 使用统一标识符
	ea.stats.TypeDistribution["standard_error"]++
}

// analyzeTags 分析标签统计（调用时已持有锁）
func (ea *ErrorAnalyzer) analyzeTags(tags map[string][]string) {
	for key, values := range tags {
		if ea.stats.TagStats[key] == nil {
			ea.stats.TagStats[key] = make(map[string]int64)
		}

		for _, value := range values {
			// 限制标签值统计数量
			if len(ea.stats.TagStats[key]) < ea.maxTagStats {
				ea.stats.TagStats[key][value]++
			}
		}
	}
}

// GetStats 获取当前统计信息（读锁保护）
func (ea *ErrorAnalyzer) GetStats() ErrorStats {
	ea.mu.RLock()
	defer ea.mu.RUnlock()

	// 返回副本，避免并发问题
	stats := *ea.stats

	// 深拷贝映射
	stats.CodeDistribution = make(map[int64]int64)
	for k, v := range ea.stats.CodeDistribution {
		stats.CodeDistribution[k] = v
	}

	stats.TypeDistribution = make(map[string]int64)
	for k, v := range ea.stats.TypeDistribution {
		stats.TypeDistribution[k] = v
	}

	stats.TagStats = make(map[string]map[string]int64)
	for k, v := range ea.stats.TagStats {
		stats.TagStats[k] = make(map[string]int64)
		for k2, v2 := range v {
			stats.TagStats[k][k2] = v2
		}
	}

	return stats
}

// Reset 重置统计信息
func (ea *ErrorAnalyzer) Reset() {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	now := time.Now()
	ea.stats = &ErrorStats{
		FirstErrorTime:   now,
		LastErrorTime:    now,
		CodeDistribution: make(map[int64]int64),
		TypeDistribution: make(map[string]int64),
		TagStats:         make(map[string]map[string]int64),
	}
}

// Summary 生成统计摘要 - 返回有用的结构化数据
func (ea *ErrorAnalyzer) Summary() AnalysisSummary {
	stats := ea.GetStats()

	// 计算时间跨度
	duration := stats.LastErrorTime.Sub(stats.FirstErrorTime)

	// 计算错误率（错误数/时间，每分钟）
	errorRate := float64(0)
	if duration.Minutes() > 0 {
		errorRate = float64(stats.TotalErrors) / duration.Minutes()
	}

	// 获取最常见的错误码（前5个）
	var topCodes []CodeStat
	for code, count := range stats.CodeDistribution {
		topCodes = append(topCodes, CodeStat{Code: code, Count: count})
	}

	// 简单排序（按count降序）
	for i := 0; i < len(topCodes)-1; i++ {
		for j := i + 1; j < len(topCodes); j++ {
			if topCodes[j].Count > topCodes[i].Count {
				topCodes[i], topCodes[j] = topCodes[j], topCodes[i]
			}
		}
	}

	// 限制为前5个
	if len(topCodes) > 5 {
		topCodes = topCodes[:5]
	}

	return AnalysisSummary{
		TotalErrors:        stats.TotalErrors,
		ErrorRate:          errorRate,
		Duration:           duration,
		TopErrorCodes:      topCodes,
		MostFrequentErrors: stats.TypeDistribution,
	}
}

// SummaryString 生成字符串摘要
func (ea *ErrorAnalyzer) SummaryString() string {
	stats := ea.GetStats()

	return fmt.Sprintf(
		"Total: %d, WithCode: %d, WithTrace: %d, Wrapped: %d, Types: %d, Codes: %d",
		stats.TotalErrors,
		stats.ErrorsWithCode,
		stats.ErrorsWithTrace,
		stats.WrappedErrors,
		len(stats.TypeDistribution),
		len(stats.CodeDistribution),
	)
}
