package statement

import (
	"strings"
)

// DeleteFrom creates delete from statement
func DeleteFrom(table string) *DeleteFromStatement {
	return &DeleteFromStatement{
		table: table,
	}
}

// DeleteFromStatement type
type DeleteFromStatement struct {
	table string
	WhereClause
}

func (stmt *DeleteFromStatement) QueryString() string {
	var b strings.Builder
	b.Grow(len(stmt.table) + 20)

	b.WriteString("delete from ")
	b.WriteString(stmt.table)

	where := stmt.WhereClause.QueryString()
	if len(where) > 0 {
		b.WriteString(" ")
		b.WriteString(where)
	}

	return b.String()
}

func (stmt *DeleteFromStatement) Args() []interface{} {
	return stmt.WhereClause.Args()
}
