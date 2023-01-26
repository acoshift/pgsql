package pgstmt

func Union(f func(b UnionStatement)) *Result {
	var st unionStmt
	f(&st)
	return newResult(build(st.make()))
}

type UnionStatement interface {
	Select(f func(b SelectStatement))
	AllSelect(f func(b SelectStatement))
	OrderBy(col string) OrderBy
	Limit(n int64)
}

type unionStmt struct {
	b       buffer
	orderBy group
	limit   *int64
}

func (st *unionStmt) Select(f func(b SelectStatement)) {
	var x selectStmt
	f(&x)

	if st.b.empty() {
		st.b.push(paren(x.make()))
	} else {
		st.b.push("union", paren(x.make()))
	}
}

func (st *unionStmt) AllSelect(f func(b SelectStatement)) {
	var x selectStmt
	f(&x)

	if st.b.empty() {
		st.b.push(paren(x.make()))
	} else {
		st.b.push("union all", paren(x.make()))
	}
}

func (st *unionStmt) OrderBy(col string) OrderBy {
	p := orderBy{
		col: col,
	}
	st.orderBy.push(&p)
	return &p
}

func (st *unionStmt) Limit(n int64) {
	st.limit = &n
}

func (st *unionStmt) make() *buffer {
	var b buffer
	b.push(&st.b)
	if !st.orderBy.empty() {
		b.push("order by", &st.orderBy)
	}
	if st.limit != nil {
		b.push("limit", *st.limit)
	}
	return &b
}
