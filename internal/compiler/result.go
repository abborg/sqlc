package compiler

import (
	"github.com/abborg/sqlc/internal/sql/catalog"
)

type Result struct {
	Catalog *catalog.Catalog
	Queries []*Query
}
