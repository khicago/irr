package reporter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/khicago/irr"
)

// Mock Logger for testing
type mockLogger struct {
	logs []logEntry
}

type logEntry struct {
	level   string
	message string
	fields  []Field
}

func (m *mockLogger) Debug(msg string, fields ...Field) {
	m.logs = append(m.logs, logEntry{level: "debug", message: msg, fields: fields})
}

func (m *mockLogger) Info(msg string, fields ...Field) {
	m.logs = append(m.logs, logEntry{level: "info", message: msg, fields: fields})
}

func (m *mockLogger) Warn(msg string, fields ...Field) {
	m.logs = append(m.logs, logEntry{level: "warn", message: msg, fields: fields})
}

func (m *mockLogger) Error(msg string, fields ...Field) {
	m.logs = append(m.logs, logEntry{level: "error", message: msg, fields: fields})
}

func (m *mockLogger) Fatal(msg string, fields ...Field) {
	m.logs = append(m.logs, logEntry{level: "fatal", message: msg, fields: fields})
}

func (m *mockLogger) hasField(key string, value interface{}) bool {
	for _, log := range m.logs {
		for _, field := range log.fields {
			if field.Key == key && field.Value == value {
				return true
			}
		}
	}
	return false
}

// Mock Alerter for testing
type mockAlerter struct {
	alerts []alertEntry
}

type alertEntry struct {
	level  ReportLevel
	err    error
	fields []Field
}

func (m *mockAlerter) SendAlert(level ReportLevel, err error, fields []Field) error {
	m.alerts = append(m.alerts, alertEntry{level: level, err: err, fields: fields})
	return nil
}

// Mock MetricsReporter for testing
type mockMetricsReporter struct {
	metrics []metricsEntry
}

type metricsEntry struct {
	level ReportLevel
	err   error
}

func (m *mockMetricsReporter) ReportErrorMetrics(level ReportLevel, err error) error {
	m.metrics = append(m.metrics, metricsEntry{level: level, err: err})
	return nil
}

func TestReportLevel_String(t *testing.T) {
	tests := []struct {
		level    ReportLevel
		expected string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarning, "warning"},
		{LevelError, "error"},
		{LevelCritical, "critical"},
		{ReportLevel(999), "unknown"},
	}

	for _, test := range tests {
		if result := test.level.String(); result != test.expected {
			t.Errorf("Expected level %d to be %s, got %s", test.level, test.expected, result)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MinLevel != LevelInfo {
		t.Errorf("Expected MinLevel to be LevelInfo, got %v", config.MinLevel)
	}

	if config.MetricsLevel != LevelWarning {
		t.Errorf("Expected MetricsLevel to be LevelWarning, got %v", config.MetricsLevel)
	}

	if config.AlertLevel != LevelError {
		t.Errorf("Expected AlertLevel to be LevelError, got %v", config.AlertLevel)
	}

	if !config.IncludeStackTrace {
		t.Error("Expected IncludeStackTrace to be true")
	}

	if config.MaxTagsPerReport != 50 {
		t.Errorf("Expected MaxTagsPerReport to be 50, got %d", config.MaxTagsPerReport)
	}
}

func TestStructuredReporter_BasicReporting(t *testing.T) {
	logger := &mockLogger{}
	config := &ReporterConfig{
		MinLevel:          LevelDebug,
		MetricsLevel:      LevelError,
		AlertLevel:        LevelCritical,
		IncludeStackTrace: true,
		MaxTagsPerReport:  10,
	}

	reporter := NewStructuredReporter(logger, config)

	// Test different report levels
	testErr := irr.New("test error").Build()

	reporter.ReportDebug(testErr)
	reporter.ReportInfo(testErr)
	reporter.ReportWarning(testErr)
	reporter.ReportError(testErr)
	reporter.ReportCritical(testErr)

	// Check that all levels were logged
	if len(logger.logs) != 5 {
		t.Errorf("Expected 5 log entries, got %d", len(logger.logs))
	}

	expectedLevels := []string{"debug", "info", "warn", "error", "fatal"}
	for i, expected := range expectedLevels {
		if logger.logs[i].level != expected {
			t.Errorf("Expected log level %s, got %s", expected, logger.logs[i].level)
		}
	}
}

func TestStructuredReporter_MinLevelFiltering(t *testing.T) {
	logger := &mockLogger{}
	config := &ReporterConfig{
		MinLevel: LevelWarning,
	}

	reporter := NewStructuredReporter(logger, config)
	testErr := irr.New("test error").Build()

	// These should be filtered out
	reporter.ReportDebug(testErr)
	reporter.ReportInfo(testErr)

	// These should be logged
	reporter.ReportWarning(testErr)
	reporter.ReportError(testErr)

	if len(logger.logs) != 2 {
		t.Errorf("Expected 2 log entries after filtering, got %d", len(logger.logs))
	}
}

func TestStructuredReporter_WithMetrics(t *testing.T) {
	logger := &mockLogger{}
	metrics := &mockMetricsReporter{}
	config := &ReporterConfig{
		MinLevel:     LevelDebug,
		MetricsLevel: LevelWarning,
	}

	reporter := NewStructuredReporter(logger, config).WithMetrics(metrics)
	testErr := irr.New("test error").Build()

	// Below metrics level - should not report metrics
	reporter.ReportInfo(testErr)

	// At/above metrics level - should report metrics
	reporter.ReportWarning(testErr)
	reporter.ReportError(testErr)

	if len(metrics.metrics) != 2 {
		t.Errorf("Expected 2 metrics reports, got %d", len(metrics.metrics))
	}
}

func TestStructuredReporter_WithAlerter(t *testing.T) {
	logger := &mockLogger{}
	alerter := &mockAlerter{}
	config := &ReporterConfig{
		MinLevel:   LevelDebug,
		AlertLevel: LevelError,
	}

	reporter := NewStructuredReporter(logger, config).WithAlerter(alerter)
	testErr := irr.New("test error").Build()

	// Below alert level - should not alert
	reporter.ReportWarning(testErr)

	// At/above alert level - should alert
	reporter.ReportError(testErr)
	reporter.ReportCritical(testErr)

	if len(alerter.alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(alerter.alerts))
	}
}

func TestStructuredReporter_IRRv2Fields(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultConfig()

	reporter := NewStructuredReporter(logger, config)

	// Create IRR v2 error with various features
	err := irr.New("test error").
		Code(404).
		Tag("module", "test").
		Tag("severity", "high").
		Trace().
		Build()

	reporter.ReportError(err)

	// Check that IRR-specific fields were logged
	if !logger.hasField("irr_version", "v2") {
		t.Error("Missing irr_version field")
	}

	if !logger.hasField("error_code", int64(404)) {
		t.Error("Missing error_code field")
	}

	if !logger.hasField("tag_module", "test") {
		t.Error("Missing tag_module field")
	}

	if !logger.hasField("tag_severity", "high") {
		t.Error("Missing tag_severity field")
	}
}

func TestStructuredReporter_ContextualReporting(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultConfig()

	reporter := NewStructuredReporter(logger, config)

	// Create context with deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := irr.New("test error").BuildWithContext(ctx)

	// StructuredReporter implements ContextualErrorReporter
	reporter.ReportWithContext(ctx, LevelError, err)

	// Check that context fields were included
	found := false
	for _, log := range logger.logs {
		for _, field := range log.fields {
			if field.Key == "context_deadline" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Missing context deadline field")
	}
}

func TestStructuredReporter_TagLimiting(t *testing.T) {
	logger := &mockLogger{}
	config := &ReporterConfig{
		MinLevel:         LevelInfo,
		MaxTagsPerReport: 2, // Limit to 2 tags
	}

	reporter := NewStructuredReporter(logger, config)

	// Create error with more tags than the limit
	builder := irr.New("test error")
	for i := 0; i < 5; i++ {
		builder = builder.Tag("key"+string(rune('0'+i)), "value")
	}
	err := builder.Build()

	reporter.ReportError(err)

	// Count tag fields (should be limited to MaxTagsPerReport)
	tagCount := 0
	for _, log := range logger.logs {
		for _, field := range log.fields {
			if strings.HasPrefix(field.Key, "tag_key") {
				tagCount++
			}
		}
	}

	if tagCount != 2 {
		t.Errorf("Expected 2 tag fields due to limit, got %d", tagCount)
	}
}

func TestNoOpReporter(t *testing.T) {
	reporter := &NoOpReporter{}
	testErr := irr.New("test error").Build()

	// These should all be no-ops and not panic
	reporter.ReportDebug(testErr)
	reporter.ReportInfo(testErr)
	reporter.ReportWarning(testErr)
	reporter.ReportError(testErr)
	reporter.ReportCritical(testErr)
}

func TestFilteringReporter(t *testing.T) {
	underlying := &mockLogger{}
	config := DefaultConfig()
	baseReporter := NewStructuredReporter(underlying, config)

	// Filter: only report errors containing "important"
	filter := func(err error) bool {
		return strings.Contains(err.Error(), "important")
	}

	reporter := NewFilteringReporter(baseReporter, filter)

	// This should be filtered out
	reporter.ReportError(irr.New("normal error").Build())

	// This should pass through
	reporter.ReportError(irr.New("important error").Build())

	if len(underlying.logs) != 1 {
		t.Errorf("Expected 1 log entry after filtering, got %d", len(underlying.logs))
	}

	if !strings.Contains(underlying.logs[0].fields[1].Value.(string), "important") {
		t.Error("Wrong error was logged after filtering")
	}
}

func TestStructuredReporter_NilError(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultConfig()
	reporter := NewStructuredReporter(logger, config)

	// Should handle nil error gracefully
	reporter.ReportError(nil)

	if len(logger.logs) != 0 {
		t.Errorf("Expected no logs for nil error, got %d", len(logger.logs))
	}
}
