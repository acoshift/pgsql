package statement

import (
	"strings"
)

// InsertInto creates insert into statement
func InsertInto(table string) *InsertIntoStatement {
	return &InsertIntoStatement{
		table: table,
	}
}

// InsertIntoStatement type
type InsertIntoStatement struct {
	table     string
	columns   []string
	values    []interface{}
	returning []string
}

func (stmt *InsertIntoStatement) Columns(cols ...string) {
	stmt.columns = append(stmt.columns, cols...)
}

func (stmt *InsertIntoStatement) Values(values ...interface{}) {
	stmt.values = append(stmt.values, values...)
}

func (stmt *InsertIntoStatement) Returning(cols ...string) {
	stmt.returning = append(stmt.returning, cols...)
}

func (stmt *InsertIntoStatement) QueryString() string {
	var b strings.Builder
	b.Grow(12 + len(stmt.table) + 2 + 10 + 1 + 50)

	b.WriteString("insert into ")
	b.WriteString(stmt.table)

	if len(stmt.columns) > 0 {
		b.WriteString(" (")
		b.WriteString(strings.Join(stmt.columns, ", "))
		b.WriteString(")")
	}

	if len(stmt.values) > 0 {
		b.WriteString(" values (")
		b.WriteString(placeHolder(1, len(stmt.values)))
		b.WriteString(")")
	}

	if len(stmt.returning) > 0 {
		b.WriteString(" returning ")
		b.WriteString(strings.Join(stmt.returning, ", "))
	}

	b.WriteString(";")

	return b.String()
}

func (stmt *InsertIntoStatement) Args() []interface{} {
	return stmt.values
}
