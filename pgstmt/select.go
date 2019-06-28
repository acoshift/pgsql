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
	Join(table string) Join
	InnerJoin(table string) Join
	FullOuterJoin(table string) Join
	LeftJoin(table string) Join
	RightJoin(table string) Join
	Where(f func(b Where))
	OrderBy(col string) OrderBy
	Limit(n int64)
	Offset(n int64)
}

type OrderBy interface {
	Asc() OrderBy
	Desc() OrderBy
	NullsFirst() OrderBy
	NullsLast() OrderBy
}

type Join interface {
	On(f func(b Where))
	Using(col ...string)
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

func (st *selectStmt) join(typ, table string) Join {
	x := join{
		typ:   typ,
		table: table,
	}
	st.joins.push(&x)
	return &x
}

func (st *selectStmt) Join(table string) Join {
	return st.join("join", table)
}

func (st *selectStmt) InnerJoin(table string) Join {
	return st.join("inner join", table)
}

func (st *selectStmt) FullOuterJoin(table string) Join {
	return st.join("full outer join", table)
}

func (st *selectStmt) LeftJoin(table string) Join {
	return st.join("left join", table)
}

func (st *selectStmt) RightJoin(table string) Join {
	return st.join("right join", table)
}

func (st *selectStmt) Where(f func(b Where)) {
	f(&st.where)
}

func (st *selectStmt) OrderBy(col string) OrderBy {
	p := orderBy{
		col: col,
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

type join struct {
	typ   string // join, inner join, full outer join, left join, right join
	table string
	using group
	on    where
}

func (st *join) On(f func(b Where)) {
	f(&st.on)
}

func (st *join) Using(col ...string) {
	var p parenGroup
	p.pushString(col...)
	st.using.push(&p)
}

func (st *join) build() []interface{} {
	var b buffer
	b.push(st.typ, st.table)
	if !st.using.empty() {
		b.push("using")
		b.push(&st.using)
	}
	if !st.on.empty() {
		b.push("on", &st.on)
	}
	return b.q
}

type orderBy struct {
	col       string
	direction string
	nulls     string
}

func (st *orderBy) Asc() OrderBy {
	st.direction = "asc"
	return st
}

func (st *orderBy) Desc() OrderBy {
	st.direction = "desc"
	return st
}

func (st *orderBy) NullsFirst() OrderBy {
	st.nulls = "first"
	return st
}

func (st *orderBy) NullsLast() OrderBy {
	st.nulls = "last"
	return st
}

func (st *orderBy) build() []interface{} {
	var b buffer
	b.push(st.col)
	if st.direction != "" {
		b.push(st.direction)
	}
	if st.nulls != "" {
		b.push("nulls", st.nulls)
	}
	return b.q
}
