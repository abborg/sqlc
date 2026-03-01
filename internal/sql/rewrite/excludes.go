package rewrite

import (
	"strings"

	"github.com/sqlc-dev/sqlc/internal/source"
	"github.com/sqlc-dev/sqlc/internal/sql/ast"
	"github.com/sqlc-dev/sqlc/internal/sql/astutils"
)

/*

type Exclude struct {
	Catalog string
	Schema  string
	Table   string
	Column  string
	param   string
}

func ParseExclude(param string) Exclude {
	return parseExcludeParam(param)
}

func parseExcludeParam(param string) Exclude {
	e := Exclude{param: param}
	parts := strings.Split(param, ".")
	switch len(parts) {
	case 1:
		e.Column = parts[0]
	case 2:
		e.Table = parts[0]
		e.Column = parts[1]
	case 3:
		e.Schema = parts[0]
		e.Table = parts[1]
		e.Column = parts[2]
	case 4:
		e.Catalog = parts[0]
		e.Schema = parts[1]
		e.Table = parts[2]
		e.Column = parts[3]
	default:
		if len(parts) >= 1 {
			e.Column = parts[len(parts)-1]
		}
	}
	return e
}

func (e Exclude) String() string {
	return e.param
}

func (e Exclude) Identifiers() []string {
	ids := []string{e.Column}
	if e.Table != "" {
		ids = append(ids, e.Table+"."+e.Column)
	}
	if e.Schema != "" {
		ids = append(ids, e.Schema+"."+e.Table+"."+e.Column)
	}
	if e.Catalog != "" {
		ids = append(ids, e.Catalog+"."+e.Schema+"."+e.Table+"."+e.Column)
	}
	return ids
}

*/

// ExcludeSet is a set of Exclude instances
type ExcludeSet []string

// Matches returns true if the given column matches any exclude identifier.
func (es ExcludeSet) Matches(col *ast.TableName, tableAlias, originalName string) bool {
	for _, excl := range es {
		parts := strings.Split(excl, ".")
		switch len(parts) {
		case 1:
			if originalName == parts[0] {
				return true
			}
		case 2:
			qualifier, name := parts[0], parts[1]
			if originalName != name {
				continue
			}
			if tableAlias == qualifier {
				return true
			}
			if col != nil && col.Name == qualifier {
				return true
			}
		case 3:
			schema, qualifier, name := parts[0], parts[1], parts[2]
			if originalName != name {
				continue
			}
			if col == nil {
				continue
			}
			if col.Schema != schema {
				continue
			}
			if tableAlias == qualifier || col.Name == qualifier {
				return true
			}
		}
	}
	return false
}

// Excludes rewrites the AST to remove sqlc.exclude(col) from TargetList and
// ReturningList. Each sqlc.exclude accepts one column identifier (string
// literal or ColumnRef). Returns the modified AST, the set of excluded
// column identifiers, and source edits to remove the exclude calls from
// the SQL text.
func Excludes(raw *ast.RawStmt, query string) (*ast.RawStmt, ExcludeSet, []source.Edit) {
	var excludes ExcludeSet
	var edits []source.Edit

	processList := func(list *ast.List, stmtStart int) {
		if list == nil || len(list.Items) == 0 {
			return
		}

		// Collect indices and info for exclude items (process backwards for safe removal)
		type excludeInfo struct {
			index int
			call  *ast.FuncCall
		}
		var toRemove []excludeInfo
		for i, item := range list.Items {
			res, ok := item.(*ast.ResTarget)
			if !ok {
				continue
			}
			call, ok := res.Val.(*ast.FuncCall)
			if !ok || !isExclude(call) {
				continue
			}
			if len(call.Args.Items) == 0 {
				continue
			}
			ident, ok := identifierFromArg(call.Args.Items[0])
			if !ok {
				continue
			}
			excludes = append(excludes, ident)
			toRemove = append(toRemove, excludeInfo{index: i, call: call})
		}

		// Remove from end first to preserve indices
		for j := len(toRemove) - 1; j >= 0; j-- {
			info := toRemove[j]
			edit := excludeEdit(stmtStart, list, info.index, info.call, query)
			if edit != nil {
				edits = append(edits, *edit)
			}
			// Remove item from list
			list.Items = append(list.Items[:info.index], list.Items[info.index+1:]...)
		}
	}

	switch stmt := raw.Stmt.(type) {
	case *ast.SelectStmt:
		processList(stmt.TargetList, raw.StmtLocation)
	case *ast.InsertStmt:
		processList(stmt.ReturningList, raw.StmtLocation)
	case *ast.UpdateStmt:
		processList(stmt.ReturningList, raw.StmtLocation)
	case *ast.DeleteStmt:
		processList(stmt.ReturningList, raw.StmtLocation)
	}

	return raw, excludes, edits
}

func isExclude(node ast.Node) bool {
	call, ok := node.(*ast.FuncCall)
	if !ok || call.Func == nil {
		return false
	}
	return call.Func.Schema == "sqlc" && call.Func.Name == "exclude"
}

func identifierFromArg(arg ast.Node) (string, bool) {
	switch n := arg.(type) {
	case *ast.A_Const:
		if n.Val != nil {
			if s, ok := n.Val.(*ast.String); ok {
				return s.Str, true
			}
		}
		return "", false
	case *ast.ColumnRef:
		if n.Fields == nil {
			return "", false
		}
		return astutils.Join(n.Fields, "."), true
	default:
		return "", false
	}
}

// extentOfExcludeCall returns the length of sqlc.exclude(...) in s, including the
// closing parenthesis. The string s should start at or before the "sqlc.exclude"
// text. Returns the length from the first "sqlc.exclude" to the matching ')'.
func extentOfExcludeCall(s string) int {
	idx := strings.Index(strings.ToLower(s), "sqlc.exclude")
	if idx < 0 {
		return 0
	}
	// Find opening paren after sqlc.exclude
	parenStart := strings.Index(s[idx:], "(")
	if parenStart < 0 {
		return len(s)
	}
	parenStart += idx
	depth := 1
	for i := parenStart + 1; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				// Include from idx through the closing paren (inclusive)
				return i - idx + 1
			}
		}
	}
	return len(s)
}

// excludeEdit creates a source edit to remove the sqlc.exclude call from the query.
func excludeEdit(stmtStart int, list *ast.List, excludedIndex int, call *ast.FuncCall, query string) *source.Edit {
	callStart := call.Location - stmtStart
	if callStart < 0 {
		callStart = call.Location
	}
	if callStart >= len(query) {
		return nil
	}

	oldFunc := func(s string) int {
		return extentOfExcludeCall(s)
	}

	isFirst := excludedIndex == 0
	isLast := excludedIndex == len(list.Items)-1

	var location int

	if isFirst && !isLast {
		// First item: remove "sqlc.exclude(x), "
		location = callStart
		oldFuncWithComma := func(s string) int {
			n := extentOfExcludeCall(s)
			// Consume trailing ", " or ","
			for i := n; i < len(s); i++ {
				if s[i] == ' ' || s[i] == '\t' {
					continue
				}
				if s[i] == ',' {
					// Include comma and optional space
					j := i + 1
					for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
						j++
					}
					return j
				}
				break
			}
			return n
		}
		return &source.Edit{
			Location: location,
			Old:      "",
			OldFunc:  oldFuncWithComma,
			New:      "",
		}
	}
	if !isFirst {
		// Not first: remove ", sqlc.exclude(x)" - find comma before.
		// Use OldFunc so Mutate computes the extent at apply time (including closing paren).
		for i := callStart - 1; i >= 0; i-- {
			if query[i] == ',' {
				location = i
				// s = query[location:], so s starts with ", sqlc.exclude(...)"
				// Return length from comma through end of sqlc.exclude(...)
				oldFuncFromComma := func(s string) int {
					idx := strings.Index(strings.ToLower(s), "sqlc.exclude")
					if idx < 0 {
						return 1
					}
					excludeLen := extentOfExcludeCall(s[idx:])
					return idx + excludeLen
				}
				return &source.Edit{
					Location: location,
					Old:      "",
					OldFunc:  oldFuncFromComma,
					New:      "",
				}
			}
			if query[i] != ' ' && query[i] != '\t' && query[i] != '\n' && query[i] != '\r' {
				break
			}
		}
	}

	// Fallback: just the exclude call
	location = callStart
	return &source.Edit{
		Location: location,
		Old:      "",
		OldFunc:  oldFunc,
		New:      "",
	}
}
