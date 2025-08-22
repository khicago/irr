package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/khicago/irr"
)

// TracingBackend 分布式跟踪后端接口
// 允许用户使用任何分布式跟踪库（OpenTracing, OpenTelemetry 等）
type TracingBackend interface {
	// Span 管理
	StartSpan(operationName string, tags map[string]interface{}) Span
	StartChildSpan(parent Span, operationName string, tags map[string]interface{}) Span

	// 从 Context 获取当前 Span（如果存在）
	SpanFromContext(ctx context.Context) Span
}

// Span 分布式跟踪 Span 接口
type Span interface {
	// 设置标签
	SetTag(key string, value interface{}) Span
	SetBaggageItem(restrictedKey, value string) Span

	// 记录日志
	LogKV(alternatingKeyValues ...interface{})
	LogFields(fields ...LogField)

	// 设置操作名称
	SetOperationName(operationName string) Span

	// 完成 Span
	Finish()
	FinishWithOptions(opts FinishOptions)

	// 获取 Span 上下文
	Context() SpanContext
}

// SpanContext Span 上下文接口
type SpanContext interface {
	TraceID() string
	SpanID() string
}

// LogField 日志字段
type LogField struct {
	Key   string
	Value interface{}
}

// FinishOptions Span 完成选项
type FinishOptions struct {
	FinishTime time.Time
}

// JaegerAnalyzer Jaeger 错误分析器
// 将错误信息记录到分布式跟踪系统中
type JaegerAnalyzer struct {
	tracer TracingBackend
	config *TracingConfig
}

// TracingConfig 分布式跟踪配置
type TracingConfig struct {
	// 是否记录错误详情
	RecordErrorDetails bool

	// 是否记录堆栈跟踪
	RecordStackTrace bool

	// 是否为错误创建独立的 Span
	CreateErrorSpan bool

	// 错误 Span 的操作名称前缀
	ErrorSpanPrefix string

	// 标签白名单
	AllowedTags []string

	// 最大堆栈深度
	MaxStackDepth int
}

// NewJaegerAnalyzer 创建 Jaeger 分析器
func NewJaegerAnalyzer(tracer TracingBackend) *JaegerAnalyzer {
	return &JaegerAnalyzer{
		tracer: tracer,
		config: &TracingConfig{
			RecordErrorDetails: true,
			RecordStackTrace:   true,
			CreateErrorSpan:    false, // 默认不创建独立 span，而是记录到现有 span
			ErrorSpanPrefix:    "error:",
			MaxStackDepth:      20,
		},
	}
}

// WithConfig 设置配置
func (ja *JaegerAnalyzer) WithConfig(config *TracingConfig) *JaegerAnalyzer {
	ja.config = config
	return ja
}

// WithErrorSpan 启用错误独立 Span 创建
func (ja *JaegerAnalyzer) WithErrorSpan(create bool) *JaegerAnalyzer {
	ja.config.CreateErrorSpan = create
	return ja
}

// WithTags 设置标签白名单
func (ja *JaegerAnalyzer) WithTags(allowedTags ...string) *JaegerAnalyzer {
	ja.config.AllowedTags = allowedTags
	return ja
}

// Analyze 分析错误并记录到 Jaeger
func (ja *JaegerAnalyzer) Analyze(err error) {
	ja.AnalyzeWithContext(context.Background(), err)
}

// AnalyzeWithContext 在指定 Context 中分析错误
func (ja *JaegerAnalyzer) AnalyzeWithContext(ctx context.Context, err error) {
	if err == nil {
		return
	}

	var span Span

	if ja.config.CreateErrorSpan {
		// 创建独立的错误 Span
		if parentSpan := ja.tracer.SpanFromContext(ctx); parentSpan != nil {
			span = ja.tracer.StartChildSpan(parentSpan, ja.config.ErrorSpanPrefix+"error_occurred", nil)
		} else {
			span = ja.tracer.StartSpan(ja.config.ErrorSpanPrefix+"error_occurred", nil)
		}
		defer span.Finish()
	} else {
		// 使用现有的 Span（如果存在）
		span = ja.tracer.SpanFromContext(ctx)
		if span == nil {
			return // 没有活动的 span，无法记录
		}
	}

	// 记录基础错误信息
	ja.recordBasicErrorInfo(span, err)

	// 如果是 IRR v2 错误，记录更详细的信息
	if irrErr, ok := err.(irr.IRR); ok {
		ja.recordIRRErrorInfo(span, irrErr)
	}
}

// recordBasicErrorInfo 记录基础错误信息
func (ja *JaegerAnalyzer) recordBasicErrorInfo(span Span, err error) {
	// 设置错误标签
	span.SetTag("error", true)
	span.SetTag("error.object", fmt.Sprintf("%T", err))
	span.SetTag("error.message", err.Error())

	// 记录错误日志
	span.LogFields(
		LogField{Key: "event", Value: "error"},
		LogField{Key: "error.message", Value: err.Error()},
		LogField{Key: "error.type", Value: fmt.Sprintf("%T", err)},
		LogField{Key: "timestamp", Value: time.Now().Unix()},
	)
}

// recordIRRErrorInfo 记录 IRR 特定信息
func (ja *JaegerAnalyzer) recordIRRErrorInfo(span Span, err irr.IRR) {
	// 设置 IRR 特定标签
	span.SetTag("irr.version", "v2")

	// 错误码信息
	if err.HasCode() {
		code := err.Code()
		span.SetTag("irr.code", code)
		span.SetTag("irr.code.severity", ja.getCodeSeverity(code))
	}

	// 错误链信息
	chainLength := ja.calculateChainLength(err)
	span.SetTag("irr.chain_length", chainLength)

	if err.Unwrap() != nil {
		span.SetTag("irr.is_wrapped", true)
		span.SetTag("irr.root_message", err.Root().Error())
	} else {
		span.SetTag("irr.is_wrapped", false)
	}

	// 堆栈跟踪信息
	if ja.config.RecordStackTrace {
		if traceable, ok := err.(irr.IRR); ok && traceable.HasStackTrace() {
			span.SetTag("irr.has_stack_trace", true)
			ja.recordStackTrace(span, traceable.StackTrace())
		} else {
			span.SetTag("irr.has_stack_trace", false)
		}
	}

	// 标签信息
	if ja.config.RecordErrorDetails {
		ja.recordErrorTags(span, err)
	}

	// Context 信息
	if contextual, ok := err.(irr.IRR); ok {
		ja.recordContextInfo(span, contextual)
	}

	// 记录详细的日志字段
	logFields := []LogField{
		{Key: "irr.version", Value: "v2"},
		{Key: "irr.has_code", Value: err.HasCode()},
		{Key: "irr.chain_length", Value: chainLength},
	}

	if err.HasCode() {
		logFields = append(logFields, LogField{Key: "irr.code", Value: err.Code()})
	}

	span.LogFields(logFields...)
}

// recordErrorTags 记录错误标签
func (ja *JaegerAnalyzer) recordErrorTags(span Span, err irr.IRR) {
	if taggable, ok := err.(irr.IRR); ok {
		tags := taggable.Tags()

		span.SetTag("irr.tags_count", len(tags))

		// 记录每个标签
		for key, values := range tags {
			// 检查标签白名单
			if !ja.isAllowedTag(key) {
				continue
			}

			// 设置标签到 span
			tagKey := "irr.tag." + key
			if len(values) == 1 {
				span.SetTag(tagKey, values[0])
			} else {
				span.SetTag(tagKey, values) // 多值标签
			}
		}

		// 记录标签日志
		span.LogFields(LogField{Key: "irr.tags", Value: tags})
	}
}

// recordStackTrace 记录堆栈跟踪
func (ja *JaegerAnalyzer) recordStackTrace(span Span, frames []irr.Frame) {
	if len(frames) == 0 {
		return
	}

	span.SetTag("irr.stack_trace.depth", len(frames))

	// 限制堆栈深度避免过多数据
	maxDepth := ja.config.MaxStackDepth
	if len(frames) > maxDepth {
		frames = frames[:maxDepth]
	}

	// 记录堆栈帧
	stackInfo := make([]map[string]interface{}, len(frames))
	for i, frame := range frames {
		stackInfo[i] = map[string]interface{}{
			"file":     frame.File,
			"line":     frame.Line,
			"function": frame.Function,
			"package":  frame.Package,
		}
	}

	span.LogFields(
		LogField{Key: "irr.stack_trace", Value: stackInfo},
		LogField{Key: "irr.stack_trace.top_function", Value: frames[0].Function},
		LogField{Key: "irr.stack_trace.top_file", Value: frames[0].File},
	)
}

// recordContextInfo 记录 Context 信息
func (ja *JaegerAnalyzer) recordContextInfo(span Span, contextual irr.IRR) {
	ctx := contextual.Context()
	if ctx == nil {
		return
	}

	// 记录 Context 状态
	if ctx.Err() != nil {
		span.SetTag("irr.context.error", ctx.Err().Error())
		span.SetTag("irr.context.cancelled", true)
	}

	// 记录超时信息
	if deadline, ok := ctx.Deadline(); ok {
		span.SetTag("irr.context.has_deadline", true)
		span.SetTag("irr.context.deadline", deadline.Format(time.RFC3339))
		span.SetTag("irr.context.remaining", time.Until(deadline).String())
	} else {
		span.SetTag("irr.context.has_deadline", false)
	}
}

// calculateChainLength 计算错误链长度
func (ja *JaegerAnalyzer) calculateChainLength(err error) int {
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

// getCodeSeverity 获取错误码严重程度
func (ja *JaegerAnalyzer) getCodeSeverity(code int64) string {
	switch {
	case code >= 1000 && code < 2000:
		return "system"
	case code >= 2000 && code < 3000:
		return "business"
	case code >= 3000 && code < 4000:
		return "api"
	case code >= 4000 && code < 5000:
		return "integration"
	case code >= 5000:
		return "security"
	case code >= 400 && code < 500:
		return "client"
	case code >= 500 && code < 600:
		return "server"
	default:
		return "unknown"
	}
}

// isAllowedTag 检查标签是否在白名单中
func (ja *JaegerAnalyzer) isAllowedTag(tagKey string) bool {
	if len(ja.config.AllowedTags) == 0 {
		return true // 没有白名单，允许所有标签
	}

	for _, allowed := range ja.config.AllowedTags {
		if tagKey == allowed {
			return true
		}
	}

	return false
}

// AnalyzeBatch 批量分析错误
func (ja *JaegerAnalyzer) AnalyzeBatch(ctx context.Context, errors []error) {
	if len(errors) == 0 {
		return
	}

	var span Span
	if ja.config.CreateErrorSpan {
		if parentSpan := ja.tracer.SpanFromContext(ctx); parentSpan != nil {
			span = ja.tracer.StartChildSpan(parentSpan, ja.config.ErrorSpanPrefix+"batch_error_analysis", map[string]interface{}{
				"batch.size": len(errors),
			})
		} else {
			span = ja.tracer.StartSpan(ja.config.ErrorSpanPrefix+"batch_error_analysis", map[string]interface{}{
				"batch.size": len(errors),
			})
		}
		defer span.Finish()
	}

	for i, err := range errors {
		if err == nil {
			continue
		}

		// 为批处理中的每个错误创建子 span
		if span != nil {
			childSpan := ja.tracer.StartChildSpan(span, "error_item", map[string]interface{}{
				"item.index": i,
			})
			ja.recordBasicErrorInfo(childSpan, err)
			if irrErr, ok := err.(irr.IRR); ok {
				ja.recordIRRErrorInfo(childSpan, irrErr)
			}
			childSpan.Finish()
		}
	}

	if span != nil {
		span.LogFields(
			LogField{Key: "batch.completed", Value: true},
			LogField{Key: "batch.error_count", Value: len(errors)},
		)
	}
}

// 示例：OpenTracing 适配器
// 用户可以根据自己使用的跟踪库实现 TracingBackend 接口

/*
type OpenTracingAdapter struct {
	tracer opentracing.Tracer
}

func NewOpenTracingAdapter(tracer opentracing.Tracer) *OpenTracingAdapter {
	return &OpenTracingAdapter{tracer: tracer}
}

func (ot *OpenTracingAdapter) StartSpan(operationName string, tags map[string]interface{}) Span {
	span := ot.tracer.StartSpan(operationName)
	for k, v := range tags {
		span.SetTag(k, v)
	}
	return &OpenTracingSpanAdapter{span: span}
}

// ... 其他方法的实现
*/
