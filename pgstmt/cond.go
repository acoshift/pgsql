package pgstmt

// Cond is the condition builder
type Cond interface {
	Op(field any, op string, value any)
	OpRaw(field any, op string, rawValue any)
	Eq(field, value any)
	EqRaw(field, rawValue any)
	Ne(field, value any)
	NeRaw(field, rawValue any)
	Lt(field, value any)
	LtRaw(field, rawValue any)
	Le(field, value any)
	LeRaw(field, rawValue any)
	Gt(field, value any)
	GtRaw(field, rawValue any)
	Ge(field, value any)
	GeRaw(field, rawValue any)
	Like(field, value any)
	LikeRaw(field, rawValue any)
	ILike(field, value any)
	ILikeRaw(field, rawValue any)
	In(field any, value ...any)
	InRaw(field any, value ...any)
	InSelect(field any, f func(b SelectStatement))
	NotIn(field any, value ...any)
	NotInRaw(field any, value ...any)
	IsNull(field any)
	IsNotNull(field any)

	Field(field any) CondOp
	Value(value any) CondOp

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

type CondOp interface {
	Op(op string) CondValue
	OpValues(op string) CondValues
	Eq() CondValue
	Ne() CondValue
	Lt() CondValue
	Le() CondValue
	Gt() CondValue
	Ge() CondValue
	Like() CondValue
	ILike() CondValue
	In() CondValues
	NotIn() CondValues
	IsNull()
	IsNotNull()
}

type CondValue interface {
	Value(value any)
	Raw(rawValue any)
	Field(field any)
}

type CondValues interface {
	Value(values ...any)
	Raw(rawValues ...any)
	Field(field any)
	Select(f func(b SelectStatement))
}

type cond struct {
	ops    parenGroup
	chain  buffer
	nested bool
}

func (st *cond) Op(field any, op string, value any) {
	var x group
	x.sep = " "
	x.push(field, op, Arg(value))
	st.ops.push(&x)
}

func (st *cond) OpRaw(field any, op string, rawValue any) {
	var x group
	x.sep = " "
	x.push(field, op, Raw(rawValue))
	st.ops.push(&x)
}

func (st *cond) Eq(field, value any) {
	st.Op(field, "=", value)
}

func (st *cond) EqRaw(field, rawValue any) {
	st.OpRaw(field, "=", rawValue)
}

func (st *cond) Ne(field, value any) {
	st.Op(field, "!=", value)
}

func (st *cond) NeRaw(field, rawValue any) {
	st.OpRaw(field, "!=", rawValue)
}

func (st *cond) Lt(field, value any) {
	st.Op(field, "<", value)
}

func (st *cond) LtRaw(field, rawValue any) {
	st.OpRaw(field, "<", rawValue)
}

func (st *cond) Le(field, value any) {
	st.Op(field, "<=", value)
}

func (st *cond) LeRaw(field, rawValue any) {
	st.OpRaw(field, "<=", rawValue)
}

func (st *cond) Gt(field, value any) {
	st.Op(field, ">", value)
}

func (st *cond) GtRaw(field, rawValue any) {
	st.OpRaw(field, ">", rawValue)
}

func (st *cond) Ge(field, value any) {
	st.Op(field, ">=", value)
}

func (st *cond) GeRaw(field, rawValue any) {
	st.OpRaw(field, ">=", rawValue)
}

func (st *cond) Like(field, value any) {
	st.Op(field, "like", value)
}

func (st *cond) LikeRaw(field, rawValue any) {
	st.OpRaw(field, "like", rawValue)
}

func (st *cond) ILike(field, value any) {
	st.Op(field, "ilike", value)
}

func (st *cond) ILikeRaw(field, rawValue any) {
	st.OpRaw(field, "ilike", rawValue)
}

func (st *cond) In(field any, value ...any) {
	var p group
	for _, v := range value {
		p.push(Arg(v))
	}

	var x group
	x.sep = " "
	x.push(field, "in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) InRaw(field any, value ...any) {
	var p group
	p.push(value...)

	var x group
	x.sep = " "
	x.push(field, "in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) InSelect(field any, f func(b SelectStatement)) {
	var x selectStmt
	f(&x)

	var p group
	p.sep = " "
	p.push(field, "in", paren(x.make()))
	st.ops.push(&p)
}

func (st *cond) NotIn(field any, value ...any) {
	var p group
	for _, v := range value {
		p.push(Arg(v))
	}

	var x group
	x.sep = " "
	x.push(field, "not in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) NotInRaw(field any, value ...any) {
	var p group
	p.push(value...)

	var x group
	x.sep = " "
	x.push(field, "not in", paren(&p))
	st.ops.push(&x)
}

func (st *cond) IsNull(field any) {
	var x group
	x.sep = " "
	x.push(field, "is null")
	st.ops.push(&x)
}

func (st *cond) IsNotNull(field any) {
	var x group
	x.sep = " "
	x.push(field, "is not null")
	st.ops.push(&x)
}

func (st *cond) Field(field any) CondOp {
	var x condOp
	x.field = field
	st.ops.push(&x)
	return &x
}

func (st *cond) Value(value any) CondOp {
	var x condOp
	x.field = Arg(value)
	st.ops.push(&x)
	return &x
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

type condOp struct {
	field  any
	op     string
	value  *condValue
	values *condValues
}

func (op *condOp) build() []any {
	var b buffer

	var x group
	x.sep = " "
	x.push(op.field, op.op)
	if op.value != nil {
		x.push(op.value.value)
	} else if op.values != nil {
		x.push(&op.values.b)
	}

	b.push(&x)

	return b.build()
}

func (op *condOp) Op(s string) CondValue {
	op.op = s
	op.value = &condValue{}
	return op.value
}

func (op *condOp) OpValues(s string) CondValues {
	op.op = s
	op.values = &condValues{}
	return op.values
}

func (op *condOp) Eq() CondValue {
	return op.Op("=")
}

func (op *condOp) Ne() CondValue {
	return op.Op("!=")
}

func (op *condOp) Lt() CondValue {
	return op.Op("<")
}

func (op *condOp) Le() CondValue {
	return op.Op("<=")
}

func (op *condOp) Gt() CondValue {
	return op.Op(">")
}

func (op *condOp) Ge() CondValue {
	return op.Op(">=")
}

func (op *condOp) Like() CondValue {
	return op.Op("like")
}

func (op *condOp) ILike() CondValue {
	return op.Op("ilike")
}

func (op *condOp) In() CondValues {
	return op.OpValues("in")
}

func (op *condOp) NotIn() CondValues {
	return op.OpValues("not in")
}

func (op *condOp) IsNull() {
	op.op = "is null"
}

func (op *condOp) IsNotNull() {
	op.op = "is not null"
}

type condValue struct {
	value any
}

func (v *condValue) make() *buffer {
	var b buffer
	b.push(v.value)
	return &b
}

func (v *condValue) Value(value any) {
	v.value = Arg(value)
}

func (v *condValue) Raw(rawValue any) {
	v.value = Raw(rawValue)
}

func (v *condValue) Field(field any) {
	v.value = Raw(field)
}

type condValues struct {
	b buffer
}

func (v *condValues) Value(value ...any) {
	var p group
	for _, x := range value {
		p.push(Arg(x))
	}
	v.b.push(paren(&p))
}

func (v *condValues) Raw(rawValue ...any) {
	var p group
	p.push(rawValue...)
	v.b.push(paren(&p))
}

func (v *condValues) Field(field any) {
	v.b.push(Raw(field))
}

func (v *condValues) Select(f func(b SelectStatement)) {
	var x selectStmt
	f(&x)

	v.b.push(paren(x.make()))
}
