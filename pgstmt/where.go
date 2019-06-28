package pgstmt

// Where is the where clause builder
type Where interface {
	Op(field, op string, value interface{})
	OpRaw(field, op string, rawValue interface{})
	Eq(field string, value interface{})
	EqRaw(field string, rawValue interface{})
	Ne(field string, value interface{})
	NeRaw(field string, rawValue interface{})
	Lt(field string, value interface{})
	LtRaw(field string, rawValue interface{})
	Le(field string, value interface{})
	LeRaw(field string, rawValue interface{})
	Gt(field string, value interface{})
	GtRaw(field string, rawValue interface{})
	Ge(field string, value interface{})
	GeRaw(field string, rawValue interface{})
	Like(field string, value interface{})
	LikeRaw(field string, rawValue interface{})
	ILike(field string, value interface{})
	ILikeRaw(field string, rawValue interface{})
	And(f func(b Where))
	Or(f func(b Where))
}

type where struct {
	ops    parenGroup
	chain  buffer
	nested bool
}

func (st *where) Op(field, op string, value interface{}) {
	var x group
	x.sep = " "
	x.push(field, op, Arg(value))
	st.ops.push(&x)
}

func (st *where) OpRaw(field, op string, rawValue interface{}) {
	st.Op(field, op, NotArg(rawValue))
}

func (st *where) Eq(field string, value interface{}) {
	st.Op(field, "=", value)
}

func (st *where) EqRaw(field string, rawValue interface{}) {
	st.Eq(field, NotArg(rawValue))
}

func (st *where) Ne(field string, value interface{}) {
	st.Op(field, "!=", value)
}

func (st *where) NeRaw(field string, rawValue interface{}) {
	st.Ne(field, NotArg(rawValue))
}

func (st *where) Lt(field string, value interface{}) {
	st.Op(field, "<", value)
}

func (st *where) LtRaw(field string, rawValue interface{}) {
	st.Lt(field, NotArg(rawValue))
}

func (st *where) Le(field string, value interface{}) {
	st.Op(field, "<=", value)
}

func (st *where) LeRaw(field string, rawValue interface{}) {
	st.Le(field, NotArg(rawValue))
}

func (st *where) Gt(field string, value interface{}) {
	st.Op(field, ">", value)
}

func (st *where) GtRaw(field string, rawValue interface{}) {
	st.Gt(field, NotArg(rawValue))
}

func (st *where) Ge(field string, value interface{}) {
	st.Op(field, ">=", value)
}

func (st *where) GeRaw(field string, rawValue interface{}) {
	st.Ge(field, NotArg(rawValue))
}

func (st *where) Like(field string, value interface{}) {
	st.Op(field, "like", value)
}

func (st *where) LikeRaw(field string, rawValue interface{}) {
	st.Like(field, NotArg(rawValue))
}

func (st *where) ILike(field string, value interface{}) {
	st.Op(field, "ilike", value)
}

func (st *where) ILikeRaw(field string, rawValue interface{}) {
	st.ILike(field, NotArg(rawValue))
}

func (st *where) And(f func(b Where)) {
	var x where
	x.ops.sep = " and "
	x.nested = true
	f(&x)

	if !x.empty() {
		st.chain.push("and")
		st.chain.push(&x)
	}
}

func (st *where) Or(f func(b Where)) {
	var x where
	x.ops.sep = " and "
	x.nested = true
	f(&x)

	if !x.empty() {
		st.chain.push("or")
		st.chain.push(&x)
	}
}

func (st *where) empty() bool {
	return st.ops.empty() /* && st.chain.empty() */
}

func (st *where) build() []interface{} {
	if st.ops.empty() {
		return nil
	}

	if st.ops.sep == "" {
		st.ops.sep = " and "
	}

	if st.nested && !st.chain.empty() {
		var b parenGroup
		b.sep = " "
		b.push(&st.ops)
		b.push(st.chain.q...)
		return []interface{}{&b}
	}

	var b buffer
	b.push(&st.ops)
	b.push(st.chain.q...)
	return b.q
}
