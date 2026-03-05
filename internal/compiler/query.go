package compiler

import (
	"github.com/abborg/sqlc/internal/metadata"
	"github.com/abborg/sqlc/internal/sql/ast"
	"github.com/abborg/sqlc/internal/sql/catalog"
)

type Function struct {
	Rel        *ast.FuncName
	ReturnType *ast.TypeName
	Outs       []*catalog.Argument
}

type Table struct {
	Rel     *ast.TableName
	Columns []*Column
}

type Column struct {
	Name         string
	OriginalName string
	DataType     string
	NotNull      bool
	Unsigned     bool
	IsArray      bool
	ArrayDims    int
	Comment      string
	Length       *int
	IsNamedParam bool
	IsFuncCall   bool

	// XXX: Figure out what PostgreSQL calls `foo.id`
	Scope      string
	Table      *ast.TableName
	TableAlias string
	Type       *ast.TypeName
	EmbedTable *ast.TableName

	IsEmbedMany bool
	IsSqlcSlice bool

	skipTableRequiredCheck bool
}

type Query struct {
	SQL      string
	Metadata metadata.Metadata
	Columns  []*Column
	Params   []Parameter

	InsertIntoTable *ast.TableName

	GroupByColumnIndex int

	// Needed for vet
	RawStmt *ast.RawStmt

	// Column identifiers to exclude from output (e.g. from sqlc.exclude)
	ExcludedColumns []string
}

type Parameter struct {
	Number int
	Column *Column
}
