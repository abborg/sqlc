package compiler

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/sqlc/internal/sql/catalog"
)

type colIdent struct {
	index int
	ids   []string
}

func flattenColumnIdentifiers(c *catalog.Catalog, columns []*Column) ([]colIdent, error) {
	var out []colIdent
	idx := 0
	for _, col := range columns {
		if col.EmbedTable != nil {
			tbl, err := c.GetTable(col.EmbedTable)
			if err != nil {
				return nil, err
			}
			scope := col.EmbedTable.Name
			if col.EmbedTable.Schema != "" {
				scope = col.EmbedTable.Schema + "." + col.EmbedTable.Name
			}
			for i, fc := range tbl.Columns {
				qual := scope + "." + fc.Name
				ids := []string{qual, fc.Name}
				if i == 0 {
					ids = append(ids, scope)
				}
				out = append(out, colIdent{index: idx, ids: ids})
				idx++
			}
		} else {
			orig := col.OriginalName
			if orig == "" {
				orig = col.Name
			}
			ids := []string{col.Name, orig}
			if col.TableAlias != "" {
				ids = append(ids, col.TableAlias+"."+orig)
			}
			if col.Table != nil && col.Table.Name != "" && col.Table.Name != col.TableAlias {
				ids = append(ids, col.Table.Name+"."+orig)
			}
			out = append(out, colIdent{index: idx, ids: ids})
			idx++
		}
	}
	return out, nil
}

func normalizeIdent(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func matchIdent(want, have string) bool {
	return normalizeIdent(want) == normalizeIdent(have)
}

func ResolveGroupByColumnIndex(c *catalog.Catalog, columns []*Column, identifier string) (int, error) {
	if identifier == "" {
		return -1, nil
	}
	flat, err := flattenColumnIdentifiers(c, columns)
	if err != nil {
		return -1, err
	}
	var matched []int
	want := normalizeIdent(identifier)
	for _, ci := range flat {
		for _, id := range ci.ids {
			if matchIdent(want, id) {
				matched = append(matched, ci.index)
				break
			}
		}
	}
	if len(matched) == 0 {
		return -1, fmt.Errorf("group-by column %q not found in query result", identifier)
	}
	if len(matched) > 1 {
		return -1, fmt.Errorf("group-by column %q is ambiguous", identifier)
	}
	return matched[0], nil
}

func (c *Compiler) resolveGroupBy(q *Query) error {
	hasEmbedMany := false
	for _, col := range q.Columns {
		if col.IsEmbedMany {
			hasEmbedMany = true
			break
		}
	}
	if !hasEmbedMany {
		return nil
	}
	ident := q.Metadata.Params["group-by"]
	if ident == "" {
		return fmt.Errorf("query uses sqlc.embed_many but @group-by is not specified")
	}
	idx, err := ResolveGroupByColumnIndex(c.catalog, q.Columns, ident)
	if err != nil {
		return err
	}
	q.GroupByColumnIndex = idx
	return nil
}
