package enrichers

import (
	"context"
	"time"

	"github.com/khicago/irr"
)

// TimeoutEnricher 超时检测丰富器
type TimeoutEnricher struct{}

// Enrich 检测超时信息
func (t *TimeoutEnricher) Enrich(ctx context.Context, err irr.IRR) irr.IRR {
	if ctx == nil {
		return err
	}

	if tagger, ok := err.(interface{ SetTag(string, string) }); ok {
		if contextErr := ctx.Err(); contextErr != nil {
			switch contextErr {
			case context.DeadlineExceeded:
				tagger.SetTag("timeout", "true")
				if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
					tagger.SetTag("deadline", deadline.Format(time.RFC3339))
				}
			case context.Canceled:
				tagger.SetTag("canceled", "true")
			}
		}
	}

	return err
}
