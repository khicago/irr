package enrichers

import (
	"context"

	"github.com/khicago/irr"
)

// DatabaseEnricher 数据库操作丰富器
type DatabaseEnricher struct {
	TableName    string
	Operation    string
	DatabaseName string
}

// Enrich 添加数据库操作信息
func (d *DatabaseEnricher) Enrich(ctx context.Context, err irr.IRR) irr.IRR {
	if tagger, ok := err.(interface{ SetTag(string, string) }); ok {
		if d.TableName != "" {
			tagger.SetTag("db_table", d.TableName)
		}
		if d.Operation != "" {
			tagger.SetTag("db_operation", d.Operation)
		}
		if d.DatabaseName != "" {
			tagger.SetTag("db_name", d.DatabaseName)
		}
	}

	return err
}
