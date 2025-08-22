package irr

import (
	"context"
)

// Enricher 接口 - 基于第一性原理的统一设计
// 负责丰富错误的上下文信息，职责单一
type Enricher interface {
	// Enrich 丰富错误的上下文信息
	// 返回丰富后的错误，不修改原始错误对象
	Enrich(ctx context.Context, err IRR) IRR
}

// 堆栈跟踪丰富器
type traceEnricher struct{}

func (te *traceEnricher) Enrich(ctx context.Context, err IRR) IRR {
	// 简化实现：如果没有堆栈跟踪则添加
	if !err.HasStackTrace() {
		// 这里可以添加实际的堆栈跟踪逻辑
		// 为了简化，暂时返回原错误
	}
	return err
}

// NewTraceEnricher 创建堆栈跟踪丰富器
func NewTraceEnricher() Enricher {
	return &traceEnricher{}
}
