package enrichers

import (
	"context"
	"testing"
	"time"

	"github.com/khicago/irr"
	"github.com/stretchr/testify/assert"
)

func TestRequestInfoEnricher_Basic(t *testing.T) {
	enricher := NewRequestInfoEnricher()
	assert.NotNil(t, enricher)
	assert.Len(t, enricher.TraceIDKeys, 3)
	assert.Len(t, enricher.UserIDKeys, 3)
	assert.Len(t, enricher.RequestIDKeys, 3)
}

func TestRequestInfoEnricher_WithContext(t *testing.T) {
	enricher := NewRequestInfoEnricher()

	// 创建带有请求信息的context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "trace-12345")
	ctx = context.WithValue(ctx, "user_id", int64(42))
	ctx = context.WithValue(ctx, "request_id", "req-67890")

	// 创建错误
	err := irr.Error("test error")

	// 应用丰富器
	enrichedErr := enricher.Enrich(ctx, err)

	// 验证标签被正确设置
	tags := enrichedErr.Tags()
	assert.Equal(t, "trace-12345", tags["trace_id"][0])
	assert.Equal(t, "42", tags["user_id"][0])
	assert.Equal(t, "req-67890", tags["request_id"][0])
}

func TestRequestInfoEnricher_WithAlternativeKeys(t *testing.T) {
	enricher := NewRequestInfoEnricher()

	// 使用备用的key名
	ctx := context.Background()
	ctx = context.WithValue(ctx, "x-trace-id", "alt-trace-12345")
	ctx = context.WithValue(ctx, "x-user-id", 99)
	ctx = context.WithValue(ctx, "x-request-id", "alt-req-67890")

	err := irr.Error("test error")
	enrichedErr := enricher.Enrich(ctx, err)

	tags := enrichedErr.Tags()
	assert.Equal(t, "alt-trace-12345", tags["trace_id"][0])
	assert.Equal(t, "99", tags["user_id"][0])
	assert.Equal(t, "alt-req-67890", tags["request_id"][0])
}

func TestRequestInfoEnricher_NilContext(t *testing.T) {
	enricher := NewRequestInfoEnricher()
	err := irr.Error("test error")

	// nil context应该返回原错误
	enrichedErr := enricher.Enrich(nil, err)
	assert.Equal(t, err, enrichedErr)
}

func TestRequestInfoEnricher_EmptyValues(t *testing.T) {
	enricher := NewRequestInfoEnricher()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "") // 空字符串
	ctx = context.WithValue(ctx, "user_id", nil) // nil值

	err := irr.Error("test error")
	enrichedErr := enricher.Enrich(ctx, err)

	tags := enrichedErr.Tags()
	assert.Empty(t, tags["trace_id"])
	assert.Empty(t, tags["user_id"])
}

func TestRequestInfoEnricher_ExtractValue(t *testing.T) {
	enricher := NewRequestInfoEnricher()

	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "test", "test"},
		{"int", 42, "42"},
		{"int64", int64(123), "123"},
		{"empty string", "", ""},
		{"nil", nil, ""},
		{"unsupported type", 3.14, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "test_key", tt.value)
			result := enricher.extractValue(ctx, "test_key")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeoutEnricher_NoTimeout(t *testing.T) {
	enricher := &TimeoutEnricher{}
	ctx := context.Background()
	err := irr.Error("test error")

	enrichedErr := enricher.Enrich(ctx, err)
	tags := enrichedErr.Tags()

	// 没有超时，不应该有相关标签
	assert.Empty(t, tags["timeout"])
	assert.Empty(t, tags["canceled"])
}

func TestTimeoutEnricher_WithDeadline(t *testing.T) {
	enricher := &TimeoutEnricher{}

	// 创建已经超时的context
	deadline := time.Now().Add(-time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// 等待context超时
	<-ctx.Done()

	err := irr.Error("timeout error")
	enrichedErr := enricher.Enrich(ctx, err)

	tags := enrichedErr.Tags()
	assert.Equal(t, "true", tags["timeout"][0])
	assert.NotEmpty(t, tags["deadline"])
}

func TestTimeoutEnricher_WithCancellation(t *testing.T) {
	enricher := &TimeoutEnricher{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	err := irr.Error("canceled error")
	enrichedErr := enricher.Enrich(ctx, err)

	tags := enrichedErr.Tags()
	assert.Equal(t, "true", tags["canceled"][0])
	assert.Empty(t, tags["timeout"])
}

func TestTimeoutEnricher_NilContext(t *testing.T) {
	enricher := &TimeoutEnricher{}
	err := irr.Error("test error")

	enrichedErr := enricher.Enrich(nil, err)
	assert.Equal(t, err, enrichedErr)
}

func TestDatabaseEnricher_Basic(t *testing.T) {
	enricher := &DatabaseEnricher{
		TableName:    "users",
		Operation:    "SELECT",
		DatabaseName: "prod_db",
	}

	ctx := context.Background()
	err := irr.Error("database error")

	enrichedErr := enricher.Enrich(ctx, err)
	tags := enrichedErr.Tags()

	assert.Equal(t, "users", tags["db_table"][0])
	assert.Equal(t, "SELECT", tags["db_operation"][0])
	assert.Equal(t, "prod_db", tags["db_name"][0])
}

func TestDatabaseEnricher_PartialFields(t *testing.T) {
	enricher := &DatabaseEnricher{
		TableName: "users",
		// Operation和DatabaseName为空
	}

	ctx := context.Background()
	err := irr.Error("database error")

	enrichedErr := enricher.Enrich(ctx, err)
	tags := enrichedErr.Tags()

	assert.Equal(t, "users", tags["db_table"][0])
	assert.Empty(t, tags["db_operation"])
	assert.Empty(t, tags["db_name"])
}

func TestDatabaseEnricher_EmptyFields(t *testing.T) {
	enricher := &DatabaseEnricher{}

	ctx := context.Background()
	err := irr.Error("database error")

	enrichedErr := enricher.Enrich(ctx, err)
	tags := enrichedErr.Tags()

	// 所有字段都为空，不应该设置任何标签
	assert.Empty(t, tags["db_table"])
	assert.Empty(t, tags["db_operation"])
	assert.Empty(t, tags["db_name"])
}

func TestEnrichersChaining(t *testing.T) {
	// 测试多个enricher的链式调用
	requestEnricher := NewRequestInfoEnricher()
	timeoutEnricher := &TimeoutEnricher{}
	dbEnricher := &DatabaseEnricher{
		TableName: "orders",
		Operation: "INSERT",
	}

	// 创建带有各种信息的context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "chain-test-123")
	ctx, cancel := context.WithCancel(ctx)
	cancel() // 模拟取消

	err := irr.Error("chained error")

	// 链式应用所有enricher
	enrichedErr := requestEnricher.Enrich(ctx, err)
	enrichedErr = timeoutEnricher.Enrich(ctx, enrichedErr)
	enrichedErr = dbEnricher.Enrich(ctx, enrichedErr)

	tags := enrichedErr.Tags()

	// 验证所有enricher都生效
	assert.Equal(t, "chain-test-123", tags["trace_id"][0])
	assert.Equal(t, "true", tags["canceled"][0])
	assert.Equal(t, "orders", tags["db_table"][0])
	assert.Equal(t, "INSERT", tags["db_operation"][0])
}
