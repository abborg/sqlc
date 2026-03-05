package ast

import "github.com/abborg/sqlc/internal/sql/format"

type List struct {
	Items []Node
}

func (n *List) Pos() int {
	return 0
}

func (n *List) Format(buf *TrackedBuffer, d format.Dialect) {
	if n == nil {
		return
	}
	buf.join(n, d, ", ")
}
