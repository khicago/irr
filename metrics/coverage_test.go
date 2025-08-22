package metrics

import (
	"context"
	"fmt"
	"testing"

	"github.com/khicago/irr"
	"github.com/stretchr/testify/assert"
)

// TestAnalyzer_SummaryString tests the summary string functionality
func TestAnalyzer_SummaryString(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	// Add some test data
	analyzer.Analyze(irr.ErrorC(400, "bad request"))
	analyzer.Analyze(irr.ErrorC(500, "server error"))

	summaryStr := analyzer.SummaryString()
	assert.NotEmpty(t, summaryStr)
	assert.Contains(t, summaryStr, "Total:")
}

// MockPrometheusMetrics for testing
type MockPrometheusMetrics struct {
	counters   map[string]int
	histograms map[string][]float64
	gauges     map[string]float64
}

func NewMockPrometheusMetrics() *MockPrometheusMetrics {
	return &MockPrometheusMetrics{
		counters:   make(map[string]int),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

func (m *MockPrometheusMetrics) IncCounter(name string, labels map[string]string) {
	key := fmt.Sprintf("%s_%v", name, labels)
	m.counters[key]++
}

func (m *MockPrometheusMetrics) AddCounter(name string, labels map[string]string, value float64) {
	key := fmt.Sprintf("%s_%v", name, labels)
	m.counters[key] += int(value)
}

func (m *MockPrometheusMetrics) ObserveHistogram(name string, labels map[string]string, value float64) {
	key := fmt.Sprintf("%s_%v", name, labels)
	m.histograms[key] = append(m.histograms[key], value)
}

func (m *MockPrometheusMetrics) SetGauge(name string, labels map[string]string, value float64) {
	key := fmt.Sprintf("%s_%v", name, labels)
	m.gauges[key] = value
}

func (m *MockPrometheusMetrics) IncGauge(name string, labels map[string]string) {
	key := fmt.Sprintf("%s_%v", name, labels)
	m.gauges[key]++
}

func (m *MockPrometheusMetrics) DecGauge(name string, labels map[string]string) {
	key := fmt.Sprintf("%s_%v", name, labels)
	m.gauges[key]--
}

// TestPrometheusAnalyzer tests Prometheus integration
func TestPrometheusAnalyzer(t *testing.T) {
	metrics := NewMockPrometheusMetrics()
	analyzer := NewPrometheusAnalyzer(metrics)

	t.Run("NewPrometheusAnalyzer creates analyzer", func(t *testing.T) {
		assert.NotNil(t, analyzer)
	})

	t.Run("WithConfig returns self", func(t *testing.T) {
		config := &PrometheusConfig{
			MetricPrefix: "test",
		}
		result := analyzer.WithConfig(config)
		assert.Equal(t, analyzer, result)
	})

	t.Run("WithPrefix returns self", func(t *testing.T) {
		result := analyzer.WithPrefix("test_prefix")
		assert.Equal(t, analyzer, result)
	})

	t.Run("WithTags returns self", func(t *testing.T) {
		result := analyzer.WithTags("env:test")
		assert.Equal(t, analyzer, result)
	})

	t.Run("Analyze handles various errors", func(t *testing.T) {
		errors := []error{
			irr.Error("simple error"),
			irr.ErrorC(404, "not found"),
			fmt.Errorf("standard error"),
			nil,
		}

		for _, err := range errors {
			assert.NotPanics(t, func() {
				analyzer.Analyze(err)
			})
		}

		// Should have recorded metrics
		assert.NotEmpty(t, metrics.counters)
	})

	t.Run("AnalyzeBatch processes multiple errors", func(t *testing.T) {
		errors := []error{
			irr.ErrorC(400, "bad request"),
			irr.ErrorC(500, "server error"),
		}

		assert.NotPanics(t, func() {
			analyzer.AnalyzeBatch(errors)
		})
	})
}

// TestPrometheusGoAdapter tests the built-in Go adapter
func TestPrometheusGoAdapter(t *testing.T) {
	adapter := NewPrometheusGoAdapter()
	assert.NotNil(t, adapter)

	labels := map[string]string{"test": "value"}

	// Test all interface methods don't panic
	assert.NotPanics(t, func() {
		adapter.IncCounter("test_counter", labels)
	})

	assert.NotPanics(t, func() {
		adapter.AddCounter("test_counter", labels, 2.0)
	})

	assert.NotPanics(t, func() {
		adapter.ObserveHistogram("test_histogram", labels, 0.5)
	})

	assert.NotPanics(t, func() {
		adapter.SetGauge("test_gauge", labels, 10.0)
	})

	assert.NotPanics(t, func() {
		adapter.IncGauge("test_gauge", labels)
	})

	assert.NotPanics(t, func() {
		adapter.DecGauge("test_gauge", labels)
	})
}

// MockTracingBackend for testing Jaeger analyzer
type MockTracingBackend struct{}

func (m *MockTracingBackend) StartSpan(operationName string, tags map[string]interface{}) Span {
	return &MockSpan{}
}

func (m *MockTracingBackend) StartChildSpan(parent Span, operationName string, tags map[string]interface{}) Span {
	return &MockSpan{}
}

func (m *MockTracingBackend) SpanFromContext(ctx context.Context) Span {
	return nil
}

// MockSpan for testing
type MockSpan struct{}

func (s *MockSpan) SetTag(key string, value interface{}) Span       { return s }
func (s *MockSpan) SetBaggageItem(restrictedKey, value string) Span { return s }
func (s *MockSpan) LogKV(alternatingKeyValues ...interface{})       {}
func (s *MockSpan) LogFields(fields ...LogField)                    {}
func (s *MockSpan) SetOperationName(operationName string) Span      { return s }
func (s *MockSpan) Finish()                                         {}
func (s *MockSpan) FinishWithOptions(opts FinishOptions)            {}
func (s *MockSpan) Context() SpanContext                            { return nil }

// TestJaegerAnalyzer tests Jaeger tracing integration
func TestJaegerAnalyzer(t *testing.T) {
	tracer := &MockTracingBackend{}
	analyzer := NewJaegerAnalyzer(tracer)

	t.Run("NewJaegerAnalyzer creates analyzer", func(t *testing.T) {
		assert.NotNil(t, analyzer)
	})

	t.Run("WithConfig returns self", func(t *testing.T) {
		config := &TracingConfig{
			RecordErrorDetails: true,
		}
		result := analyzer.WithConfig(config)
		assert.Equal(t, analyzer, result)
	})

	t.Run("WithErrorSpan returns self", func(t *testing.T) {
		result := analyzer.WithErrorSpan(true)
		assert.Equal(t, analyzer, result)
	})

	t.Run("WithTags returns self", func(t *testing.T) {
		result := analyzer.WithTags("service:test")
		assert.Equal(t, analyzer, result)
	})

	t.Run("Analyze handles errors", func(t *testing.T) {
		assert.NotPanics(t, func() {
			analyzer.Analyze(irr.Error("test error"))
		})
	})

	t.Run("AnalyzeWithContext handles errors", func(t *testing.T) {
		ctx := context.Background()
		assert.NotPanics(t, func() {
			analyzer.AnalyzeWithContext(ctx, irr.Error("test error"))
		})
	})

	t.Run("AnalyzeBatch processes multiple errors", func(t *testing.T) {
		ctx := context.Background()
		errors := []error{
			irr.Error("error1"),
			irr.Error("error2"),
		}
		assert.NotPanics(t, func() {
			analyzer.AnalyzeBatch(ctx, errors)
		})
	})
}

// ExamplePrometheusAnalyzer demonstrates Prometheus integration
func ExamplePrometheusAnalyzer() {
	metrics := NewMockPrometheusMetrics()
	analyzer := NewPrometheusAnalyzer(metrics).
		WithPrefix("myapp").
		WithTags("env:production")

	// Analyze an error
	err := irr.ErrorC(500, "database connection failed")
	analyzer.Analyze(err)

	fmt.Println("Prometheus metrics recorded")

	// Output:
	// Prometheus metrics recorded
}

// ExampleJaegerAnalyzer demonstrates Jaeger tracing integration
func ExampleJaegerAnalyzer() {
	tracer := &MockTracingBackend{}
	analyzer := NewJaegerAnalyzer(tracer).
		WithErrorSpan(true).
		WithTags("service:user-api")

	// Trace error with context
	ctx := context.Background()
	err := irr.ErrorC(404, "user not found")
	analyzer.AnalyzeWithContext(ctx, err)

	fmt.Println("Error traced successfully")

	// Output:
	// Error traced successfully
}

// BenchmarkAnalyzers benchmarks different analyzers
func BenchmarkAnalyzers(b *testing.B) {
	b.Run("ErrorAnalyzer", func(b *testing.B) {
		analyzer := NewErrorAnalyzer()
		err := irr.ErrorC(500, "benchmark error")
		for i := 0; i < b.N; i++ {
			analyzer.Analyze(err)
		}
	})

	b.Run("PrometheusAnalyzer", func(b *testing.B) {
		metrics := NewMockPrometheusMetrics()
		analyzer := NewPrometheusAnalyzer(metrics)
		err := irr.ErrorC(500, "benchmark error")
		for i := 0; i < b.N; i++ {
			analyzer.Analyze(err)
		}
	})

	b.Run("JaegerAnalyzer", func(b *testing.B) {
		tracer := &MockTracingBackend{}
		analyzer := NewJaegerAnalyzer(tracer)
		err := irr.ErrorC(500, "benchmark error")
		for i := 0; i < b.N; i++ {
			analyzer.Analyze(err)
		}
	})
}
