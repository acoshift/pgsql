package pgstmt

func Insert(f func(b *InsertBuilder)) (string, []interface{}) {
	var b InsertBuilder
	b.push("insert")
	f(&b)

	return b.build()
}

type InsertBuilder struct {
	builder

	columns   parenGroup
	values    group
	returning group
}

func (b *InsertBuilder) Into(table string) {
	b.push("into", table)
}

func (b *InsertBuilder) Columns(col ...string) {
	b.columns.pushString(col...)
}

func (b *InsertBuilder) Value(value ...interface{}) {
	var x parenGroup
	for _, v := range value {
		x.push(arg(v))
	}
	b.values.push(x)
}

func (b *InsertBuilder) Returning(field ...string) {
	b.returning.pushString(field...)
}

func (b *InsertBuilder) build() (string, []interface{}) {
	if !b.columns.empty() {
		b.push(b.columns)
	}
	if !b.values.empty() {
		b.push("values")
		b.push(b.values)
	}
	if !b.returning.empty() {
		b.push("returning")
		b.push(b.returning)
	}
	return b.builder.build()
}
