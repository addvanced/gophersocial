package store

import (
	"fmt"
	"strings"
)

type Query struct {
	b      strings.Builder
	params []any
}

func (q *Query) Query(s string) {
	q.b.WriteString(s)
}

func (q *Query) Param(val any) {
	q.b.WriteString(fmt.Sprintf("$%d", len(q.params)+1))
	q.params = append(q.params, val)
}

func (q *Query) GetQuery() string {
	return q.b.String()
}

func (q *Query) GetParams() []any {
	return q.params
}
