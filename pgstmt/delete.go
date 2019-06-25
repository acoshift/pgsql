package pgstmt

func Delete(f func(b *DeleteBuilder)) *Result {
	var b DeleteBuilder
	b.push("delete")
	f(&b)

	return newResult(b.build())
}

type DeleteBuilder struct {
	builder

	returning group
}

func (b *DeleteBuilder) From(table string) {
	b.push("from", table)
}

func (b *DeleteBuilder) Where(f func(b *WhereBuilder)) {
	var x WhereBuilder
	x.ops.sep = " and "
	f(&x)

	if !x.ops.empty() {
		b.push("where")
		b.push(&x)
	}
}

func (b *DeleteBuilder) Returning(field ...string) {
	b.returning.pushString(field...)
}

func (b *DeleteBuilder) build() (string, []interface{}) {
	if !b.returning.empty() {
		b.push("returning")
		b.push(&b.returning)
	}
	return b.builder.build()
}
