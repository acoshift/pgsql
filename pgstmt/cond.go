package pgstmt

// Cond is the condition builder
type Cond interface {
	Op(field, op string, value any)
	OpRaw(field, op string, rawValue any)
	Eq(field string, value any)
	EqRaw(field string, rawValue any)
	Ne(field string, value any)
	NeRaw(field string, rawValue any)
	Lt(field string, value any)
	LtRaw(field string, rawValue any)
	Le(field string, value any)
	LeRaw(field string, rawValue any)
	Gt(field string, value any)
	GtRaw(field string, rawValue any)
	Ge(field string, value any)
	GeRaw(field string, rawValue any)
	Like(field string, value any)
	LikeRaw(field string, rawValue any)
	ILike(field string, value any)
	ILikeRaw(field string, rawValue any)
	In(field string, value ...any)
	InRaw(field string, value ...any)
	InSelect(field string, f func(b SelectStatement))
	NotIn(field string, value ...any)
	NotInRaw(field string, value ...any)
	IsNull(field string)
	IsNotNull(field string)
	Raw(sql string)
	Not(f func(b Cond))
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

func (st *cond) Op(field, op string, value any) {
	var x group
	x.sep = " "
	x.push(field, op, Arg(value))
	st.ops.push(&x)
}

func (st *cond) OpRaw(field, op string, rawValue any) {
	var x group
	x.sep = " "
	x.push(field, op, Raw(rawValue))
	st.ops.push(&x)
}

func (st *cond) Eq(field string, value any) {
	st.Op(field, "=", value)
}

func (st *cond) EqRaw(field string, rawValue any) {
	st.OpRaw(field, "=", rawValue)
}

func (st *cond) Ne(field string, value any) {
	st.Op(field, "!=", value)
}

func (st *cond) NeRaw(field string, rawValue any) {
	st.OpRaw(field, "!=", rawValue)
}

func (st *cond) Lt(field string, value any) {
	st.Op(field, "<", value)
}

func (st *cond) LtRaw(field string, rawValue any) {
	st.OpRaw(field, "<", rawValue)
}

func (st *cond) Le(field string, value any) {
	st.Op(field, "<=", value)
}

func (st *cond) LeRaw(field string, rawValue any) {
	st.OpRaw(field, "<=", rawValue)
}

func (st *cond) Gt(field string, value any) {
	st.Op(field, ">", value)
}

func (st *cond) GtRaw(field string, rawValue any) {
	st.OpRaw(field, ">", rawValue)
}

func (st *cond) Ge(field string, value any) {
	st.Op(field, ">=", value)
}

func (st *cond) GeRaw(field string, rawValue any) {
	st.OpRaw(field, ">=", rawValue)
}

func (st *cond) Like(field string, value any) {
	st.Op(field, "like", value)
}

func (st *cond) LikeRaw(field string, rawValue any) {
	st.OpRaw(field, "like", rawValue)
}

func (st *cond) ILike(field string, value any) {
	st.Op(field, "ilike", value)
}

func (st *cond) ILikeRaw(field string, rawValue any) {
	st.OpRaw(field, "ilike", rawValue)
}

func (st *cond) In(field string, value ...any) {
	var p group
	for _, v := range value {
		p.push(Arg(v))
	}

	var x group
	x.sep = " "
	x.push(field, "in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) InRaw(field string, value ...any) {
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

func (st *cond) NotIn(field string, value ...any) {
	var p group
	for _, v := range value {
		p.push(Arg(v))
	}

	var x group
	x.sep = " "
	x.push(field, "not in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) NotInRaw(field string, value ...any) {
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

func (st *cond) Not(b func(b Cond)) {
	var x cond
	x.ops.sep = " and "
	x.nested = true
	b(&x)

	if !x.empty() {
		st.ops.push(withParen(" ", "not", &x))
	}
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

func (st *cond) build() []any {
	if st.empty() {
		return nil
	}

	if st.ops.empty() {
		st.chain.popFront()

		if len(st.chain.q) > 1 {
			var b parenGroup
			b.sep = " "
			b.push(st.chain.q...)
			return []any{&b}
		}

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
		return []any{&b}
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
