package pgstmt

// Delete builds delete statement
func Delete(f func(b DeleteStatement)) *Result {
	var st deleteStmt
	f(&st)

	var b buffer
	b.push("delete")
	b.push("from", st.from)
	if !st.where.empty() {
		b.push("where")
		b.push(st.where.build()...)
	}
	if !st.returning.empty() {
		b.push("returning")
		b.push(&st.returning)
	}

	return newResult(build(&b))
}

type DeleteStatement interface {
	From(table string)
	Where(f func(b Cond))
	Returning(col ...string)
}

type deleteStmt struct {
	from      string
	where     cond
	returning group
}

func (st *deleteStmt) From(table string) {
	st.from = table
}

func (st *deleteStmt) Where(f func(b Cond)) {
	f(&st.where)
}

func (st *deleteStmt) Returning(col ...string) {
	st.returning.pushString(col...)
}
