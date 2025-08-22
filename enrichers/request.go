// Package enrichers 提供官方的 Enricher 实现
package enrichers

import (
	"context"
	"strconv"

	"github.com/khicago/irr"
)

// RequestInfoEnricher 请求信息丰富器
type RequestInfoEnricher struct {
	TraceIDKeys   []string
	UserIDKeys    []string
	RequestIDKeys []string
}

// NewRequestInfoEnricher 创建默认的请求信息丰富器
func NewRequestInfoEnricher() *RequestInfoEnricher {
	return &RequestInfoEnricher{
		TraceIDKeys:   []string{"trace-id", "trace_id", "x-trace-id"},
		UserIDKeys:    []string{"user-id", "user_id", "x-user-id"},
		RequestIDKeys: []string{"request-id", "request_id", "x-request-id"},
	}
}

// Enrich 实现请求信息的提取和丰富
func (r *RequestInfoEnricher) Enrich(ctx context.Context, err irr.IRR) irr.IRR {
	if ctx == nil {
		return err
	}

	if tagger, ok := err.(interface{ SetTag(string, string) }); ok {
		if traceID := r.extractValue(ctx, r.TraceIDKeys...); traceID != "" {
			tagger.SetTag("trace_id", traceID)
		}
		if userID := r.extractValue(ctx, r.UserIDKeys...); userID != "" {
			tagger.SetTag("user_id", userID)
		}
		if requestID := r.extractValue(ctx, r.RequestIDKeys...); requestID != "" {
			tagger.SetTag("request_id", requestID)
		}
	}

	return err
}

func (r *RequestInfoEnricher) extractValue(ctx context.Context, keys ...string) string {
	for _, key := range keys {
		if value := ctx.Value(key); value != nil {
			switch v := value.(type) {
			case string:
				if v != "" {
					return v
				}
			case int:
				return strconv.Itoa(v)
			case int64:
				return strconv.FormatInt(v, 10)
			}
		}
	}
	return ""
}
