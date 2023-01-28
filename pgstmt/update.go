package pgstmt

// Update builds update statement
func Update(f func(b UpdateStatement)) *Result {
	var st updateStmt
	f(&st)
	return newResult(build(st.make()))
}

type UpdateStatement interface {
	Table(table string)
	Set(col ...string) Set
	From(table ...string)
	Join(table string) Join
	InnerJoin(table string) Join
	FullOuterJoin(table string) Join
	LeftJoin(table string) Join
	RightJoin(table string) Join
	Where(f func(b Cond))
	WhereCurrentOf(cursor string)
	Returning(col ...string)
}

type Set interface {
	To(value ...any)
	ToRaw(rawValue ...any)
	Select(f func(b SelectStatement))
}

type updateStmt struct {
	table          string
	sets           group
	from           group
	joins          buffer
	where          cond
	whereCurrentOf string
	returning      group
}

func (st *updateStmt) Table(table string) {
	st.table = table
}

func (st *updateStmt) Set(col ...string) Set {
	var x set
	x.col.pushString(col...)
	st.sets.push(&x)
	return &x
}

func (st *updateStmt) From(table ...string) {
	st.from.pushString(table...)
}

func (st *updateStmt) join(typ, table string) Join {
	var b buffer
	b.push(table)
	x := join{
		typ:   typ,
		table: &b,
	}
	st.joins.push(&x)
	return &x
}

func (st *updateStmt) Join(table string) Join {
	return st.join("join", table)
}

func (st *updateStmt) InnerJoin(table string) Join {
	return st.join("inner join", table)
}

func (st *updateStmt) FullOuterJoin(table string) Join {
	return st.join("full outer join", table)
}

func (st *updateStmt) LeftJoin(table string) Join {
	return st.join("left join", table)
}

func (st *updateStmt) RightJoin(table string) Join {
	return st.join("right join", table)
}

func (st *updateStmt) Where(f func(b Cond)) {
	f(&st.where)
}

func (st *updateStmt) WhereCurrentOf(cursor string) {
	st.whereCurrentOf = cursor
}

func (st *updateStmt) Returning(col ...string) {
	st.returning.pushString(col...)
}

func (st *updateStmt) make() *buffer {
	var b buffer
	b.push("update")
	if st.table != "" {
		b.push(st.table)
	}
	if !st.sets.empty() {
		b.push("set", &st.sets)
	}
	if !st.from.empty() {
		b.push("from", &st.from)
	}
	if !st.joins.empty() {
		b.push(&st.joins)
	}
	if !st.where.empty() {
		b.push("where", &st.where)
	}
	if st.whereCurrentOf != "" {
		b.push("where current of", st.whereCurrentOf)
	}
	if !st.returning.empty() {
		b.push("returning", &st.returning)
	}
	return &b
}

type set struct {
	col group
	to  group
}

func (st *set) To(value ...any) {
	for _, v := range value {
		st.to.push(Arg(v))
	}
}

func (st *set) ToRaw(rawValue ...any) {
	st.to.push(rawValue...)
}

func (st *set) Select(f func(b SelectStatement)) {
	var x selectStmt
	f(&x)
	st.to.push(paren(x.make()))
}

func (st *set) build() []any {
	var b buffer
	if len(st.col.q) > 1 {
		b.push(paren(&st.col))
	} else {
		b.push(&st.col)
	}
	b.push("=")
	if len(st.to.q) > 1 {
		var p parenGroup
		p.prefix = "row"
		p.push(&st.to)
		b.push(&p)
	} else {
		b.push(&st.to)
	}
	return b.q
}
