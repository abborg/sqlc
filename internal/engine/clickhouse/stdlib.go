package clickhouse

import (
	"github.com/abborg/sqlc/internal/sql/catalog"
)

func defaultSchema(name string) *catalog.Schema {
	return &catalog.Schema{Name: name}
}
