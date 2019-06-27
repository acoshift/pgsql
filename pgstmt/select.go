package pgstmt

// Select builds select statement
func Select(f func(b SelectStatement)) *Result {
	var st selectStmt
	f(&st)

	var b builder
	b.push("select")
	if !st.columns.empty() {
		b.push(&st.columns)
	}
	if st.from != "" {
		b.push("from")
		b.push(st.from)
	}
	if !st.where.empty() {
		b.push("where")
		b.push(st.where.build()...)
	}
	if !st.order.empty() {
		b.push("order by")
		b.push(&st.order)
	}

	return newResult(b.build())
}

// SelectStatement is the select statement builder
type SelectStatement interface {
	Columns(col ...string)
	From(sql string)
	Where(f func(b Where))
	OrderBy(col string, direction string) OrderBy
}

type selectStmt struct {
	columns group
	from    string
	where   where
	order   group
	nulls   string // first, last
}

func (st *selectStmt) Columns(col ...string) {
	st.columns.pushString(col...)
}

func (st *selectStmt) From(sql string) {
	st.from = sql
}

func (st *selectStmt) Where(f func(b Where)) {
	f(&st.where)
}

func (st *selectStmt) OrderBy(col string, op string) OrderBy {
	p := orderBy{
		col: col,
		op:  op,
	}
	st.order.push(&p)
	return &p
}

type OrderBy interface {
	NullsFirst()
	NullsLast()
}

type orderBy struct {
	col   string
	op    string
	nulls string
}

func (st *orderBy) NullsFirst() {
	st.nulls = "first"
}

func (st *orderBy) NullsLast() {
	st.nulls = "last"
}

func (st *orderBy) build() []interface{} {
	var b builder
	b.push(st.col)
	if st.op != "" {
		b.push(st.op)
	}
	if st.nulls != "" {
		b.push("nulls")
		b.push(st.nulls)
	}
	return b.q
}
