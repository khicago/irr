package metrics

import (
	"strconv"
	"time"

	"github.com/khicago/irr"
)

// PrometheusMetrics Prometheus 指标接口
// 允许用户使用任何 Prometheus 客户端库
type PrometheusMetrics interface {
	// Counter 指标
	IncCounter(name string, labels map[string]string)
	AddCounter(name string, labels map[string]string, value float64)

	// Histogram 指标
	ObserveHistogram(name string, labels map[string]string, value float64)

	// Gauge 指标
	SetGauge(name string, labels map[string]string, value float64)
	IncGauge(name string, labels map[string]string)
	DecGauge(name string, labels map[string]string)
}

// PrometheusAnalyzer Prometheus 错误分析器
// 将错误统计信息导出到 Prometheus
type PrometheusAnalyzer struct {
	metrics PrometheusMetrics
	config  *PrometheusConfig
}

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	// 指标名称前缀
	MetricPrefix string

	// 是否记录错误详情
	RecordErrorDetails bool

	// 是否记录标签信息
	RecordTags bool

	// 标签白名单（如果为空则记录所有标签）
	AllowedTags []string

	// 最大标签值数量（防止高基数问题）
	MaxLabelValues int
}

// NewPrometheusAnalyzer 创建 Prometheus 分析器
func NewPrometheusAnalyzer(metrics PrometheusMetrics) *PrometheusAnalyzer {
	return &PrometheusAnalyzer{
		metrics: metrics,
		config: &PrometheusConfig{
			MetricPrefix:       "irr_",
			RecordErrorDetails: true,
			RecordTags:         true,
			MaxLabelValues:     100,
		},
	}
}

// WithConfig 设置配置
func (pa *PrometheusAnalyzer) WithConfig(config *PrometheusConfig) *PrometheusAnalyzer {
	pa.config = config
	return pa
}

// WithPrefix 设置指标前缀
func (pa *PrometheusAnalyzer) WithPrefix(prefix string) *PrometheusAnalyzer {
	pa.config.MetricPrefix = prefix
	return pa
}

// WithTags 启用标签记录并设置白名单
func (pa *PrometheusAnalyzer) WithTags(allowedTags ...string) *PrometheusAnalyzer {
	pa.config.RecordTags = true
	pa.config.AllowedTags = allowedTags
	return pa
}

// Analyze 分析错误并导出到 Prometheus
func (pa *PrometheusAnalyzer) Analyze(err error) {
	if err == nil {
		return
	}

	startTime := time.Now()

	// 基础错误计数
	pa.recordBasicMetrics(err)

	// 如果是 IRR v2 错误，记录更详细的信息
	if irrErr, ok := err.(irr.IRR); ok {
		pa.recordIRRMetrics(irrErr)
	}

	// 记录处理耗时
	duration := time.Since(startTime).Seconds()
	pa.metrics.ObserveHistogram(
		pa.config.MetricPrefix+"error_analysis_duration_seconds",
		map[string]string{"type": "analyze"},
		duration,
	)
}

// recordBasicMetrics 记录基础指标
func (pa *PrometheusAnalyzer) recordBasicMetrics(err error) {
	labels := map[string]string{
		"type": "standard_error",
	}

	// 尝试识别错误类型
	if _, ok := err.(irr.IRR); ok {
		labels["type"] = "irr_v2"
	}

	pa.metrics.IncCounter(pa.config.MetricPrefix+"errors_total", labels)
}

// recordIRRMetrics 记录 IRR 特定指标
func (pa *PrometheusAnalyzer) recordIRRMetrics(err irr.IRR) {
	// 基础标签
	labels := map[string]string{
		"type": "irr_v2",
	}

	// 错误码指标
	if err.HasCode() {
		code := err.Code()
		codeLabels := map[string]string{
			"code":     strconv.FormatInt(code, 10),
			"severity": pa.getCodeSeverity(code),
		}

		pa.metrics.IncCounter(pa.config.MetricPrefix+"errors_with_code_total", codeLabels)

		// 按错误码分组的计数
		pa.metrics.IncCounter(pa.config.MetricPrefix+"error_codes_total", map[string]string{
			"code": strconv.FormatInt(code, 10),
		})

		labels["has_code"] = "true"
	} else {
		labels["has_code"] = "false"
	}

	// 堆栈跟踪指标
	if traceable, ok := err.(irr.IRR); ok && traceable.HasStackTrace() {
		pa.metrics.IncCounter(pa.config.MetricPrefix+"errors_with_trace_total", labels)
		labels["has_trace"] = "true"
	} else {
		labels["has_trace"] = "false"
	}

	// 包装错误指标
	if err.Unwrap() != nil {
		pa.metrics.IncCounter(pa.config.MetricPrefix+"wrapped_errors_total", labels)
		labels["is_wrapped"] = "true"
	} else {
		labels["is_wrapped"] = "false"
	}

	// 主要错误计数器（带完整标签）
	pa.metrics.IncCounter(pa.config.MetricPrefix+"irr_errors_total", labels)

	// 标签统计
	if pa.config.RecordTags && pa.config.RecordErrorDetails {
		pa.recordTagMetrics(err)
	}
}

// recordTagMetrics 记录标签相关指标
func (pa *PrometheusAnalyzer) recordTagMetrics(err irr.IRR) {
	if taggable, ok := err.(irr.IRR); ok {
		tags := taggable.Tags()

		// 记录标签数量
		pa.metrics.SetGauge(pa.config.MetricPrefix+"error_tags_count", map[string]string{}, float64(len(tags)))

		// 记录每个标签的值分布
		for key, values := range tags {
			// 检查标签白名单
			if !pa.isAllowedTag(key) {
				continue
			}

			tagLabels := map[string]string{
				"tag_key": key,
			}

			for _, value := range values {
				// 防止高基数问题
				if len(value) > 50 {
					value = value[:47] + "..." // 截断长值
				}

				valueLabels := map[string]string{
					"tag_key":   key,
					"tag_value": value,
				}

				pa.metrics.IncCounter(pa.config.MetricPrefix+"error_tag_values_total", valueLabels)
			}

			// 记录该标签的值数量
			pa.metrics.SetGauge(pa.config.MetricPrefix+"error_tag_value_count", tagLabels, float64(len(values)))
		}
	}
}

// getCodeSeverity 根据错误码判断严重程度
func (pa *PrometheusAnalyzer) getCodeSeverity(code int64) string {
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
func (pa *PrometheusAnalyzer) isAllowedTag(tagKey string) bool {
	if len(pa.config.AllowedTags) == 0 {
		return true // 没有白名单，允许所有标签
	}

	for _, allowed := range pa.config.AllowedTags {
		if tagKey == allowed {
			return true
		}
	}

	return false
}

// AnalyzeBatch 批量分析错误（提高性能）
func (pa *PrometheusAnalyzer) AnalyzeBatch(errors []error) {
	if len(errors) == 0 {
		return
	}

	startTime := time.Now()

	for _, err := range errors {
		if err != nil {
			pa.Analyze(err)
		}
	}

	// 记录批处理指标
	duration := time.Since(startTime).Seconds()
	pa.metrics.ObserveHistogram(
		pa.config.MetricPrefix+"batch_analysis_duration_seconds",
		map[string]string{
			"batch_size": strconv.Itoa(len(errors)),
		},
		duration,
	)

	pa.metrics.AddCounter(
		pa.config.MetricPrefix+"batch_operations_total",
		map[string]string{"operation": "analyze"},
		float64(len(errors)),
	)
}

// 示例：标准 Prometheus 客户端适配器
// 用户可以根据自己使用的 Prometheus 库实现这个接口

// 示例：标准 Prometheus 客户端适配器接口定义
// 用户需要根据自己使用的 Prometheus 库实现这些接口

// CounterVec Prometheus Counter 向量接口
type CounterVec interface {
	WithLabelValues(lvs ...string) Counter
	With(labels map[string]string) Counter
}

// Counter Prometheus Counter 接口
type Counter interface {
	Inc()
	Add(float64)
}

// HistogramVec Prometheus Histogram 向量接口
type HistogramVec interface {
	WithLabelValues(lvs ...string) Histogram
	With(labels map[string]string) Histogram
}

// Histogram Prometheus Histogram 接口
type Histogram interface {
	Observe(float64)
}

// GaugeVec Prometheus Gauge 向量接口
type GaugeVec interface {
	WithLabelValues(lvs ...string) Gauge
	With(labels map[string]string) Gauge
}

// Gauge Prometheus Gauge 接口
type Gauge interface {
	Set(float64)
	Inc()
	Dec()
}

// PrometheusGoAdapter prometheus/client_golang 适配器示例
// 注意：这只是示例代码，用户需要根据实际的 Prometheus 库进行调整
type PrometheusGoAdapter struct {
	counterVecs   map[string]CounterVec   // 接口类型，不是指针
	histogramVecs map[string]HistogramVec // 接口类型，不是指针
	gaugeVecs     map[string]GaugeVec     // 接口类型，不是指针
}

// NewPrometheusGoAdapter 创建 Prometheus 适配器
func NewPrometheusGoAdapter() *PrometheusGoAdapter {
	return &PrometheusGoAdapter{
		counterVecs:   make(map[string]CounterVec),
		histogramVecs: make(map[string]HistogramVec),
		gaugeVecs:     make(map[string]GaugeVec),
	}
}

// 示例实现（用户需要根据实际 Prometheus 库调整）
func (pga *PrometheusGoAdapter) IncCounter(name string, labels map[string]string) {
	if counterVec, exists := pga.counterVecs[name]; exists {
		counterVec.With(labels).Inc()
	}
}

func (pga *PrometheusGoAdapter) AddCounter(name string, labels map[string]string, value float64) {
	if counterVec, exists := pga.counterVecs[name]; exists {
		counterVec.With(labels).Add(value)
	}
}

func (pga *PrometheusGoAdapter) ObserveHistogram(name string, labels map[string]string, value float64) {
	if histogramVec, exists := pga.histogramVecs[name]; exists {
		histogramVec.With(labels).Observe(value)
	}
}

func (pga *PrometheusGoAdapter) SetGauge(name string, labels map[string]string, value float64) {
	if gaugeVec, exists := pga.gaugeVecs[name]; exists {
		gaugeVec.With(labels).Set(value)
	}
}

func (pga *PrometheusGoAdapter) IncGauge(name string, labels map[string]string) {
	if gaugeVec, exists := pga.gaugeVecs[name]; exists {
		gaugeVec.With(labels).Inc()
	}
}

func (pga *PrometheusGoAdapter) DecGauge(name string, labels map[string]string) {
	if gaugeVec, exists := pga.gaugeVecs[name]; exists {
		gaugeVec.With(labels).Dec()
	}
}
