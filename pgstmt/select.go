package pgstmt

// Select builds select statement
func Select(f func(b SelectStatement)) *Result {
	var st selectStmt
	f(&st)
	return newResult(build(st.make()))
}

// SelectStatement is the select statement builder
type SelectStatement interface {
	Columns(col ...string)
	ColumnSelect(f func(b SelectStatement), as string)
	From(table ...string)
	FromSelect(f func(b SelectStatement), as string)
	Join(table string)
	JoinOn(table string, on func(b Where))
	InnerJoin(table string)
	InnerJoinOn(table string, on func(b Where))
	FullOuterJoin(table string)
	FullOuterJoinOn(table string, on func(b Where))
	LeftJoin(table string)
	LeftJoinOn(table string, on func(b Where))
	RightJoin(table string)
	RightJoinOn(table string, on func(b Where))
	Where(f func(b Where))
	OrderBy(col string, direction string) OrderBy
	Limit(n int64)
	Offset(n int64)
}

type selectStmt struct {
	columns group
	from    group
	joins   buffer
	where   where
	order   group
	nulls   string // first, last
	limit   *int64
	offset  *int64
}

func (st *selectStmt) Columns(col ...string) {
	st.columns.pushString(col...)
}

func (st *selectStmt) ColumnSelect(f func(b SelectStatement), as string) {
	var x selectStmt
	f(&x)

	var b buffer
	var p parenGroup
	p.push(x.make())
	b.push(&p)
	if as != "" {
		b.push(as)
	}
	st.columns.push(&b)
}

func (st *selectStmt) From(table ...string) {
	st.from.pushString(table...)
}

func (st *selectStmt) FromSelect(f func(b SelectStatement), as string) {
	var x selectStmt
	f(&x)

	var b buffer
	var p parenGroup
	p.push(x.make())
	b.push(&p)
	if as != "" {
		b.push(as)
	}
	st.from.push(&b)
}

func (st *selectStmt) join(join, table string, on func(b Where)) {
	st.joins.push(join, table)
	if on != nil {
		var x where
		on(&x)
		st.joins.push("on", &x)
	}
}

func (st *selectStmt) Join(table string) {
	st.join("join", table, nil)
}

func (st *selectStmt) JoinOn(table string, on func(b Where)) {
	st.join("join", table, on)
}

func (st *selectStmt) InnerJoin(table string) {
	st.join("inner join", table, nil)
}

func (st *selectStmt) InnerJoinOn(table string, on func(b Where)) {
	st.join("inner join", table, on)
}

func (st *selectStmt) FullOuterJoin(table string) {
	st.join("full outer join", table, nil)
}

func (st *selectStmt) FullOuterJoinOn(table string, on func(b Where)) {
	st.join("full outer join", table, on)
}

func (st *selectStmt) LeftJoin(table string) {
	st.join("left join", table, nil)
}

func (st *selectStmt) LeftJoinOn(table string, on func(b Where)) {
	st.join("left join", table, on)
}

func (st *selectStmt) RightJoin(table string) {
	st.join("right join", table, nil)
}

func (st *selectStmt) RightJoinOn(table string, on func(b Where)) {
	st.join("right join", table, on)
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

func (st *selectStmt) Limit(n int64) {
	st.limit = &n
}

func (st *selectStmt) Offset(n int64) {
	st.offset = &n
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
	var b buffer
	b.push(st.col)
	if st.op != "" {
		b.push(st.op)
	}
	if st.nulls != "" {
		b.push("nulls", st.nulls)
	}
	return b.q
}

func (st *selectStmt) make() *buffer {
	var b buffer
	b.push("select")
	if !st.columns.empty() {
		b.push(&st.columns)
	}
	if !st.from.empty() {
		st.from.sep = ", "
		b.push("from", &st.from)

		if !st.joins.empty() {
			b.push(st.joins.q...)
		}
	}
	if !st.where.empty() {
		b.push("where", &st.where)
	}
	if !st.order.empty() {
		b.push("order by", &st.order)
	}
	if st.limit != nil {
		b.push("limit", *st.limit)
	}
	if st.offset != nil {
		b.push("offset", *st.offset)
	}

	return &b
}
