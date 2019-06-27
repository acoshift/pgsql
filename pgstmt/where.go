package pgstmt

// Where is the where clause builder
type Where interface {
	Op(field, op string, value interface{})
	Eq(field string, value interface{})
	Ne(field string, value interface{})
	Lt(field string, value interface{})
	Le(field string, value interface{})
	Gt(field string, value interface{})
	Ge(field string, value interface{})
	Like(field string, value interface{})
	ILike(field string, value interface{})
	Or(f func(b Where))
}

type where struct {
	ops   parenGroup
	chain builder
}

func (st *where) Op(field, op string, value interface{}) {
	var x parenGroup
	x.sep = " "
	x.push(field, op, arg(value))
	st.ops.push(&x)
}

func (st *where) Eq(field string, value interface{}) {
	st.Op(field, "=", value)
}

func (st *where) Ne(field string, value interface{}) {
	st.Op(field, "!=", value)
}

func (st *where) Lt(field string, value interface{}) {
	st.Op(field, "<", value)
}

func (st *where) Le(field string, value interface{}) {
	st.Op(field, "<=", value)
}

func (st *where) Gt(field string, value interface{}) {
	st.Op(field, ">", value)
}

func (st *where) Ge(field string, value interface{}) {
	st.Op(field, ">=", value)
}

func (st *where) Like(field string, value interface{}) {
	st.Op(field, "like", value)
}

func (st *where) ILike(field string, value interface{}) {
	st.Op(field, "ilike", value)
}

func (st *where) Or(f func(b Where)) {
	var x where
	x.ops.sep = " and "
	f(&x)

	if !x.ops.empty() {
		st.chain.push("or")
		st.chain.push(&x.ops)
	}
}

func (st *where) empty() bool {
	return st.ops.empty() && len(st.chain.q) == 0
}

func (st *where) build() []interface{} {
	if st.ops.sep == "" {
		st.ops.sep = " and "
	}

	var builder builder
	if !st.ops.empty() {
		builder.push(&st.ops)
	}
	builder.push(st.chain.q...)
	return builder.q
}
