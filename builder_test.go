package irr

import (
	"context"
	"testing"
)

// testEnricher 实现Enricher接口用于测试
type testEnricher struct {
	tagKey   string
	tagValue string
}

func (te *testEnricher) Enrich(ctx context.Context, err IRR) IRR {
	return err.Tag(te.tagKey, te.tagValue)
}

func TestErrorBuilder_Basic(t *testing.T) {
	// 测试基础错误构建
	err := New("test error").
		Code(404).
		Tag("module", "test").
		Build()

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}

	if err.Code() != 404 {
		t.Errorf("Expected code 404, got %d", err.Code())
	}

	tags := err.Tags()
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags["module"][0] != "test" {
		t.Errorf("Expected tag value 'test', got %s", tags["module"][0])
	}
}

func TestErrorBuilder_Wrap(t *testing.T) {
	// 测试错误包装
	cause := New("original error").Build()
	wrappedErr := Wraps(cause, "wrapped: %s", "context").
		Code(500).
		Build()

	if wrappedErr.Error() != "wrapped: context, original error" {
		t.Errorf("Expected 'wrapped: context, original error', got %s", wrappedErr.Error())
	}

	if wrappedErr.Code() != 500 {
		t.Errorf("Expected code 500, got %d", wrappedErr.Code())
	}

	if wrappedErr.Unwrap() != cause {
		t.Error("Expected wrapped error to unwrap to cause")
	}
}

func TestErrorBuilder_WithEnricher(t *testing.T) {
	// 测试使用丰富器
	enricher := &testEnricher{
		tagKey:   "enriched",
		tagValue: "true",
	}

	err := New("test error").
		WithEnricher(enricher).
		Build()

	tags := err.Tags()
	if tags["enriched"][0] != "true" {
		t.Errorf("Expected enriched tag 'true', got %s", tags["enriched"][0])
	}
}

func TestErrorBuilder_Trace(t *testing.T) {
	// 测试堆栈跟踪
	err := New("trace error").
		Trace().
		Build()

	if !err.HasStackTrace() {
		t.Error("Expected error to have stack trace")
	}

	trace := err.StackTrace()
	if len(trace) == 0 {
		t.Error("Expected non-empty stack trace")
	}
}

func TestErrorBuilder_BuildWithContext(t *testing.T) {
	// 测试使用上下文构建
	ctx := context.WithValue(context.Background(), "key", "value")

	err := New("context error").
		BuildWithContext(ctx)

	if err.Context() != ctx {
		t.Error("Expected error to preserve context")
	}
}

func TestErrorBuilder_ChainedOperations(t *testing.T) {
	// 测试链式操作
	err := New("chained error").
		Code(404).
		Tag("module", "test").
		Tag("severity", "high").
		Trace().
		Build()

	// 验证所有属性
	if err.Error() != "chained error" {
		t.Errorf("Expected 'chained error', got %s", err.Error())
	}

	if err.Code() != 404 {
		t.Errorf("Expected code 404, got %d", err.Code())
	}

	if !err.HasStackTrace() {
		t.Error("Expected error to have stack trace")
	}

	tags := err.Tags()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}
