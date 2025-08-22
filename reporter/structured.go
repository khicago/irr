package reporter

import (
	"context"
	"fmt"
	"time"

	"github.com/khicago/irr"
)

// StructuredReporter 结构化报告器实现
// 既记录日志又上报监控指标，解决中间过程错误报告问题
type StructuredReporter struct {
	logger  Logger
	config  *ReporterConfig
	alerter Alerter
	metrics MetricsReporter
}

// NewStructuredReporter 创建结构化报告器
func NewStructuredReporter(logger Logger, config *ReporterConfig) *StructuredReporter {
	if config == nil {
		config = DefaultConfig()
	}

	return &StructuredReporter{
		logger: logger,
		config: config,
	}
}

// WithAlerter 设置告警器
func (sr *StructuredReporter) WithAlerter(alerter Alerter) *StructuredReporter {
	sr.alerter = alerter
	return sr
}

// WithMetrics 设置指标报告器
func (sr *StructuredReporter) WithMetrics(metrics MetricsReporter) *StructuredReporter {
	sr.metrics = metrics
	return sr
}

// ReportDebug 报告调试信息
func (sr *StructuredReporter) ReportDebug(err error) {
	sr.report(LevelDebug, err, nil)
}

// ReportInfo 报告一般信息
func (sr *StructuredReporter) ReportInfo(err error) {
	sr.report(LevelInfo, err, nil)
}

// ReportWarning 报告警告
func (sr *StructuredReporter) ReportWarning(err error) {
	sr.report(LevelWarning, err, nil)
}

// ReportError 报告错误
func (sr *StructuredReporter) ReportError(err error) {
	sr.report(LevelError, err, nil)
}

// ReportCritical 报告严重错误
func (sr *StructuredReporter) ReportCritical(err error) {
	sr.report(LevelCritical, err, nil)
}

// ReportWithContext 带 Context 的报告
func (sr *StructuredReporter) ReportWithContext(ctx context.Context, level ReportLevel, err error) {
	sr.report(level, err, ctx)
}

// report 内部报告方法
func (sr *StructuredReporter) report(level ReportLevel, err error, ctx context.Context) {
	if err == nil {
		return
	}

	// 检查最小级别
	if level < sr.config.MinLevel {
		return
	}

	// 提取错误信息
	fields := sr.extractFields(err, ctx)

	// 记录日志
	sr.logError(level, err, fields)

	// 上报指标
	if level >= sr.config.MetricsLevel && sr.metrics != nil {
		sr.metrics.ReportErrorMetrics(level, err)
	}

	// 触发告警
	if level >= sr.config.AlertLevel && sr.alerter != nil {
		sr.alerter.SendAlert(level, err, fields)
	}
}

// extractFields 提取错误相关字段
func (sr *StructuredReporter) extractFields(err error, ctx context.Context) []Field {
	fields := make([]Field, 0, 10)

	// 基础字段
	fields = append(fields, Field{Key: "timestamp", Value: time.Now().Unix()})
	fields = append(fields, Field{Key: "error_message", Value: err.Error()})
	fields = append(fields, Field{Key: "error_type", Value: fmt.Sprintf("%T", err)})

	// IRR 特定信息
	if irrErr, ok := err.(irr.IRR); ok {
		fields = append(fields, Field{Key: "irr_version", Value: "v2"})

		// 错误码
		if irrErr.HasCode() {
			fields = append(fields, Field{Key: "error_code", Value: irrErr.Code()})
		}

		// 标签信息
		if taggable, ok := irrErr.(irr.IRR); ok {
			tags := taggable.Tags()
			fields = append(fields, Field{Key: "tag_count", Value: len(tags)})

			// 限制标签数量，防止高基数
			tagCount := 0
			for key, values := range tags {
				if tagCount >= sr.config.MaxTagsPerReport {
					break
				}

				tagKey := "tag_" + key
				if len(values) == 1 {
					fields = append(fields, Field{Key: tagKey, Value: values[0]})
				} else {
					fields = append(fields, Field{Key: tagKey, Value: values})
				}
				tagCount++
			}
		}

		// 堆栈跟踪
		if sr.config.IncludeStackTrace {
			if traceable, ok := irrErr.(irr.IRR); ok && traceable.HasStackTrace() {
				trace := traceable.StackTrace()
				fields = append(fields, Field{Key: "stack_trace_depth", Value: len(trace)})
				if len(trace) > 0 {
					fields = append(fields, Field{Key: "stack_trace_top", Value: trace[0].Function})
				}
			}
		}

		// 错误链长度
		chainLength := sr.calculateChainLength(err)
		fields = append(fields, Field{Key: "error_chain_length", Value: chainLength})

		// Context 信息
		if contextual, ok := irrErr.(irr.IRR); ok {
			errCtx := contextual.Context()
			if errCtx != nil && errCtx.Err() != nil {
				fields = append(fields, Field{Key: "context_error", Value: errCtx.Err().Error()})
			}
		}
	}

	// Context 额外信息
	if ctx != nil {
		if ctx.Err() != nil {
			fields = append(fields, Field{Key: "current_context_error", Value: ctx.Err().Error()})
		}

		if deadline, ok := ctx.Deadline(); ok {
			fields = append(fields, Field{Key: "context_deadline", Value: deadline.Format(time.RFC3339)})
			fields = append(fields, Field{Key: "time_remaining", Value: time.Until(deadline).String()})
		}
	}

	return fields
}

// logError 记录错误日志
func (sr *StructuredReporter) logError(level ReportLevel, err error, fields []Field) {
	if sr.logger == nil {
		return
	}

	message := fmt.Sprintf("Error reported at %s level", level.String())

	switch level {
	case LevelDebug:
		sr.logger.Debug(message, fields...)
	case LevelInfo:
		sr.logger.Info(message, fields...)
	case LevelWarning:
		sr.logger.Warn(message, fields...)
	case LevelError:
		sr.logger.Error(message, fields...)
	case LevelCritical:
		sr.logger.Fatal(message, fields...)
	}
}

// calculateChainLength 计算错误链长度
func (sr *StructuredReporter) calculateChainLength(err error) int {
	length := 1
	current := err

	for current != nil {
		if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			current = unwrapper.Unwrap()
			if current != nil {
				length++
			}
		} else {
			break
		}
	}

	return length
}

// NoOpReporter 空操作报告器，用于测试或禁用报告
type NoOpReporter struct{}

func (n *NoOpReporter) ReportDebug(err error)    {}
func (n *NoOpReporter) ReportInfo(err error)     {}
func (n *NoOpReporter) ReportWarning(err error)  {}
func (n *NoOpReporter) ReportError(err error)    {}
func (n *NoOpReporter) ReportCritical(err error) {}

// FilteringReporter 过滤报告器，只报告满足条件的错误
type FilteringReporter struct {
	underlying ErrorReporter
	filter     func(error) bool
}

// NewFilteringReporter 创建过滤报告器
func NewFilteringReporter(underlying ErrorReporter, filter func(error) bool) *FilteringReporter {
	return &FilteringReporter{
		underlying: underlying,
		filter:     filter,
	}
}

func (fr *FilteringReporter) ReportDebug(err error) {
	if fr.filter(err) {
		fr.underlying.ReportDebug(err)
	}
}

func (fr *FilteringReporter) ReportInfo(err error) {
	if fr.filter(err) {
		fr.underlying.ReportInfo(err)
	}
}

func (fr *FilteringReporter) ReportWarning(err error) {
	if fr.filter(err) {
		fr.underlying.ReportWarning(err)
	}
}

func (fr *FilteringReporter) ReportError(err error) {
	if fr.filter(err) {
		fr.underlying.ReportError(err)
	}
}

func (fr *FilteringReporter) ReportCritical(err error) {
	if fr.filter(err) {
		fr.underlying.ReportCritical(err)
	}
}
