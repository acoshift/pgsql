package pgstmt

// Cond is the condition builder
type Cond interface {
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
	In(field string, value ...interface{})
	InRaw(field string, value ...interface{})
	InSelect(field string, f func(b SelectStatement))
	NotIn(field string, value ...interface{})
	NotInRaw(field string, value ...interface{})
	IsNull(field string)
	IsNotNull(field string)
	Raw(sql string)
	And(f func(b Cond))
	Or(f func(b Cond))
	Mode() CondMode
}

type CondMode interface {
	And()
	Or()
}

type cond struct {
	ops    parenGroup
	chain  buffer
	nested bool
}

func (st *cond) Op(field, op string, value interface{}) {
	var x group
	x.sep = " "
	x.push(field, op, Arg(value))
	st.ops.push(&x)
}

func (st *cond) OpRaw(field, op string, rawValue interface{}) {
	var x group
	x.sep = " "
	x.push(field, op, rawValue)
	st.ops.push(&x)
}

func (st *cond) Eq(field string, value interface{}) {
	st.Op(field, "=", value)
}

func (st *cond) EqRaw(field string, rawValue interface{}) {
	st.OpRaw(field, "=", rawValue)
}

func (st *cond) Ne(field string, value interface{}) {
	st.Op(field, "!=", value)
}

func (st *cond) NeRaw(field string, rawValue interface{}) {
	st.OpRaw(field, "!=", rawValue)
}

func (st *cond) Lt(field string, value interface{}) {
	st.Op(field, "<", value)
}

func (st *cond) LtRaw(field string, rawValue interface{}) {
	st.OpRaw(field, "<", rawValue)
}

func (st *cond) Le(field string, value interface{}) {
	st.Op(field, "<=", value)
}

func (st *cond) LeRaw(field string, rawValue interface{}) {
	st.OpRaw(field, "<=", rawValue)
}

func (st *cond) Gt(field string, value interface{}) {
	st.Op(field, ">", value)
}

func (st *cond) GtRaw(field string, rawValue interface{}) {
	st.OpRaw(field, ">", rawValue)
}

func (st *cond) Ge(field string, value interface{}) {
	st.Op(field, ">=", value)
}

func (st *cond) GeRaw(field string, rawValue interface{}) {
	st.OpRaw(field, ">=", rawValue)
}

func (st *cond) Like(field string, value interface{}) {
	st.Op(field, "like", value)
}

func (st *cond) LikeRaw(field string, rawValue interface{}) {
	st.OpRaw(field, "like", rawValue)
}

func (st *cond) ILike(field string, value interface{}) {
	st.Op(field, "ilike", value)
}

func (st *cond) ILikeRaw(field string, rawValue interface{}) {
	st.OpRaw(field, "ilike", rawValue)
}

func (st *cond) In(field string, value ...interface{}) {
	var p group
	for _, v := range value {
		p.push(Arg(v))
	}

	var x group
	x.sep = " "
	x.push(field, "in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) InRaw(field string, value ...interface{}) {
	var p group
	p.push(value...)

	var x group
	x.sep = " "
	x.push(field, "in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) InSelect(field string, f func(b SelectStatement)) {
	var x selectStmt
	f(&x)

	var p group
	p.sep = " "
	p.push(field, "in", paren(x.make()))
	st.ops.push(&p)
}

func (st *cond) NotIn(field string, value ...interface{}) {
	var p group
	for _, v := range value {
		p.push(Arg(v))
	}

	var x group
	x.sep = " "
	x.push(field, "not in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) NotInRaw(field string, value ...interface{}) {
	var p group
	p.push(value...)

	var x group
	x.sep = " "
	x.push(field, "not in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) IsNull(field string) {
	st.ops.push(field + " is null")
}

func (st *cond) IsNotNull(field string) {
	st.ops.push(field + " is not null")
}

func (st *cond) Raw(sql string) {
	st.ops.push(sql)
}

func (st *cond) And(f func(b Cond)) {
	var x cond
	x.ops.sep = " and "
	x.nested = true
	f(&x)

	if !x.empty() {
		st.chain.push("and")
		st.chain.push(&x)
	}
}

func (st *cond) Or(f func(b Cond)) {
	var x cond
	x.ops.sep = " and "
	x.nested = true
	f(&x)

	if !x.empty() {
		st.chain.push("or")
		st.chain.push(&x)
	}
}

func (st *cond) Mode() CondMode {
	return &condMode{st}
}

func (st *cond) empty() bool {
	return st.ops.empty() && st.chain.empty()
}

func (st *cond) build() []interface{} {
	if st.empty() {
		return nil
	}

	if st.ops.empty() {
		st.chain.popFront()
		return st.chain.q
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

type condMode struct {
	cond *cond
}

func (mode *condMode) And() {
	mode.cond.ops.sep = " and "
}

func (mode *condMode) Or() {
	mode.cond.ops.sep = " or "
}
