// Package reporter 提供分级错误报告系统
// 解决中间过程日志记录问题（如cache miss、降级处理等）
package reporter

import (
	"context"
)

// ReportLevel 报告级别
type ReportLevel int

const (
	LevelDebug ReportLevel = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelCritical
)

// String 返回级别名称
func (l ReportLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarning:
		return "warning"
	case LevelError:
		return "error"
	case LevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ErrorReporter 分级错误报告接口
// 用于解决中间过程需要记录日志但不终止流程的场景
type ErrorReporter interface {
	// ReportDebug 调试信息，通常只在开发环境记录
	ReportDebug(err error)

	// ReportInfo 一般信息，如操作成功但有提示
	ReportInfo(err error)

	// ReportWarning 警告，如cache miss、降级处理等
	ReportWarning(err error)

	// ReportError 错误，但系统可恢复
	ReportError(err error)

	// ReportCritical 严重错误，需要立即关注
	ReportCritical(err error)
}

// ContextualErrorReporter 支持 Context 的报告器
// 可以从 Context 中提取更多信息
type ContextualErrorReporter interface {
	ErrorReporter

	// ReportWithContext 带 Context 的报告方法
	ReportWithContext(ctx context.Context, level ReportLevel, err error)
}

// ReporterConfig 报告器配置
type ReporterConfig struct {
	// MinLevel 最小报告级别，低于此级别的报告会被忽略
	MinLevel ReportLevel

	// MetricsLevel 只有达到此级别才上报metrics
	MetricsLevel ReportLevel

	// AlertLevel 只有达到此级别才触发告警
	AlertLevel ReportLevel

	// IncludeStackTrace 是否包含堆栈跟踪
	IncludeStackTrace bool

	// MaxTagsPerReport 每个报告最大标签数量（防止高基数）
	MaxTagsPerReport int
}

// DefaultConfig 默认配置
func DefaultConfig() *ReporterConfig {
	return &ReporterConfig{
		MinLevel:          LevelInfo,
		MetricsLevel:      LevelWarning,
		AlertLevel:        LevelError,
		IncludeStackTrace: true,
		MaxTagsPerReport:  50,
	}
}

// Logger 通用日志接口
// 允许用户使用任何日志库（zap, logrus等）
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Alerter 告警接口
// 允许用户集成任何告警系统
type Alerter interface {
	SendAlert(level ReportLevel, err error, fields []Field) error
}

// MetricsReporter 指标报告接口
// 用于将错误报告集成到监控系统
type MetricsReporter interface {
	ReportErrorMetrics(level ReportLevel, err error) error
}
