package postgresql

import (
	"github.com/abborg/sqlc/internal/sql/catalog"
)

func pgTemp() *catalog.Schema {
	return &catalog.Schema{Name: "pg_temp"}
}
