package pgstmt

type WhereBuilder struct {
	builder

	ops parenGroup
}

func (b *WhereBuilder) Op(field, op string, value interface{}) {
	var x parenGroup
	x.sep = " "
	x.push(field, op, arg(value))
	b.ops.push(&x)
}

func (b *WhereBuilder) Eq(field string, value interface{}) {
	b.Op(field, "=", value)
}

func (b *WhereBuilder) Ne(field string, value interface{}) {
	b.Op(field, "!=", value)
}

func (b *WhereBuilder) Lt(field string, value interface{}) {
	b.Op(field, "<", value)
}

func (b *WhereBuilder) Le(field string, value interface{}) {
	b.Op(field, "<=", value)
}

func (b *WhereBuilder) Gt(field string, value interface{}) {
	b.Op(field, ">", value)
}

func (b *WhereBuilder) Ge(field string, value interface{}) {
	b.Op(field, ">=", value)
}

func (b *WhereBuilder) Like(field string, value interface{}) {
	b.Op(field, "like", value)
}

func (b *WhereBuilder) ILike(field string, value interface{}) {
	b.Op(field, "ilike", value)
}

func (b *WhereBuilder) Or(f func(b *WhereBuilder)) {
	var x WhereBuilder
	x.ops.sep = " and "
	f(&x)

	if !x.ops.empty() {
		b.push("or")
		b.push(&x.ops)
	}
}

func (b *WhereBuilder) extract() []interface{} {
	b.pushFirst(&b.ops)
	return b.builder.q
}
