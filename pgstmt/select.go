package pgstmt

// Select builds select statement
func Select(f func(b SelectStatement)) *Result {
	var st selectStmt
	f(&st)
	return newResult(build(st.make()))
}

// SelectStatement is the select statement builder
type SelectStatement interface {
	Distinct() Distinct
	Columns(col ...any)
	ColumnSelect(f func(b SelectStatement), as string)
	From(table ...string)
	FromSelect(f func(b SelectStatement), as string)
	FromValues(f func(b Values), as string)

	Join(table string) Join
	InnerJoin(table string) Join
	FullOuterJoin(table string) Join
	LeftJoin(table string) Join
	RightJoin(table string) Join

	JoinSelect(f func(b SelectStatement), as string) Join
	InnerJoinSelect(f func(b SelectStatement), as string) Join
	FullOuterJoinSelect(f func(b SelectStatement), as string) Join
	LeftJoinSelect(f func(b SelectStatement), as string) Join
	RightJoinSelect(f func(b SelectStatement), as string) Join

	JoinLateralSelect(f func(b SelectStatement), as string) Join
	InnerJoinLateralSelect(f func(b SelectStatement), as string) Join
	FullOuterJoinLateralSelect(f func(b SelectStatement), as string) Join
	LeftJoinLateralSelect(f func(b SelectStatement), as string) Join
	RightJoinLateralSelect(f func(b SelectStatement), as string) Join

	JoinUnion(f func(b UnionStatement), as string) Join
	InnerJoinUnion(f func(b UnionStatement), as string) Join
	FullOuterJoinUnion(f func(b UnionStatement), as string) Join
	LeftJoinUnion(f func(b UnionStatement), as string) Join
	RightJoinUnion(f func(b UnionStatement), as string) Join

	Where(f func(b Cond))
	GroupBy(col ...string)
	Having(f func(b Cond))
	OrderBy(col string) OrderBy
	Limit(n int64)
	Offset(n int64)
}

type Distinct interface {
	On(col ...string)
}

type Values interface {
	Value(value ...any)
	Values(values ...any)
}

type OrderBy interface {
	Asc() OrderBy
	Desc() OrderBy
	NullsFirst() OrderBy
	NullsLast() OrderBy
}

type Join interface {
	On(f func(b Cond))
	Using(col ...string)
}

type selectStmt struct {
	distinct *distinct
	columns  group
	from     group
	joins    buffer
	where    cond
	groupBy  group
	having   cond
	orderBy  group
	limit    *int64
	offset   *int64
}

func (st *selectStmt) Distinct() Distinct {
	st.distinct = &distinct{}
	return st.distinct
}

func (st *selectStmt) Columns(col ...any) {
	st.columns.push(col...)
}

func (st *selectStmt) ColumnSelect(f func(b SelectStatement), as string) {
	var x selectStmt
	f(&x)

	var b buffer
	b.push(paren(x.make()))
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
	b.push(paren(x.make()))
	if as != "" {
		b.push(as)
	}
	st.from.push(&b)
}

func (st *selectStmt) FromValues(f func(b Values), as string) {
	var x values
	f(&x)

	if x.empty() {
		return
	}

	st.from.push(withGroup(" ",
		withGroup(" ",
			withParen(" ",
				"values",
				withGroup(", ", x.q...),
			),
		),
		as,
	))
}

func (st *selectStmt) join(typ, table string) Join {
	var b buffer
	b.push(table)
	x := join{
		typ:   typ,
		table: &b,
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

func (st *selectStmt) joinSelect(typ string, f func(b SelectStatement), as string) Join {
	var x selectStmt
	f(&x)

	var b buffer
	b.push(paren(x.make()))
	if as != "" {
		b.push(as)
	}

	j := join{
		typ:   typ,
		table: &b,
	}
	st.joins.push(&j)
	return &j
}

func (st *selectStmt) JoinSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("join", f, as)
}

func (st *selectStmt) InnerJoinSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("inner join", f, as)
}

func (st *selectStmt) FullOuterJoinSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("full outer join", f, as)
}

func (st *selectStmt) LeftJoinSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("left join", f, as)
}

func (st *selectStmt) RightJoinSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("right join", f, as)
}

func (st *selectStmt) JoinLateralSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("join lateral", f, as)
}

func (st *selectStmt) InnerJoinLateralSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("inner join lateral", f, as)
}

func (st *selectStmt) FullOuterJoinLateralSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("full outer join lateral", f, as)
}

func (st *selectStmt) LeftJoinLateralSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("left join lateral", f, as)
}

func (st *selectStmt) RightJoinLateralSelect(f func(b SelectStatement), as string) Join {
	return st.joinSelect("right join lateral", f, as)
}

func (st *selectStmt) joinUnion(typ string, f func(b UnionStatement), as string) Join {
	var x unionStmt
	f(&x)

	var b buffer
	b.push(paren(x.make()))
	if as != "" {
		b.push(as)
	}

	j := join{
		typ:   typ,
		table: &b,
	}
	st.joins.push(&j)
	return &j
}

func (st *selectStmt) JoinUnion(f func(b UnionStatement), as string) Join {
	return st.joinUnion("join", f, as)
}

func (st *selectStmt) InnerJoinUnion(f func(b UnionStatement), as string) Join {
	return st.joinUnion("inner join", f, as)
}

func (st *selectStmt) FullOuterJoinUnion(f func(b UnionStatement), as string) Join {
	return st.joinUnion("full outer join", f, as)
}

func (st *selectStmt) LeftJoinUnion(f func(b UnionStatement), as string) Join {
	return st.joinUnion("left join", f, as)
}

func (st *selectStmt) RightJoinUnion(f func(b UnionStatement), as string) Join {
	return st.joinUnion("right join", f, as)
}

func (st *selectStmt) Where(f func(b Cond)) {
	f(&st.where)
}

func (st *selectStmt) GroupBy(col ...string) {
	st.groupBy.pushString(col...)
}

func (st *selectStmt) Having(f func(b Cond)) {
	f(&st.having)
}

func (st *selectStmt) OrderBy(col string) OrderBy {
	p := orderBy{
		col: col,
	}
	st.orderBy.push(&p)
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
	if st.distinct != nil {
		b.push("distinct")

		if !st.distinct.columns.empty() {
			b.push("on")
			b.push(&st.distinct.columns)
		}
	}
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
	if !st.groupBy.empty() {
		b.push("group by", paren(&st.groupBy))
	}
	if !st.having.empty() {
		b.push("having", &st.having)
	}
	if !st.orderBy.empty() {
		b.push("order by", &st.orderBy)
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
	table builder
	using group
	on    cond
}

func (st *join) On(f func(b Cond)) {
	f(&st.on)
}

func (st *join) Using(col ...string) {
	st.using.push(parenString(col...))
}

func (st *join) build() []any {
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

func (st *orderBy) build() []any {
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

type values struct {
	group
}

func (st *values) Value(value ...any) {
	var x parenGroup
	for _, v := range value {
		x.push(Arg(v))
	}
	st.push(&x)
}

func (st *values) Values(values ...any) {
	for _, value := range values {
		st.Value(value)
	}
}

type distinct struct {
	columns parenGroup
}

func (st *distinct) On(col ...string) {
	st.columns.pushString(col...)
}
