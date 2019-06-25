package pgstmt

func Select(f func(b *SelectBuilder)) (string, []interface{}) {
	var b SelectBuilder
	b.push("select")
	f(&b)

	return b.build()
}

type SelectBuilder struct {
	builder

	columns group
	from    string
	where   WhereBuilder
	order   group
}

func (b *SelectBuilder) Columns(col ...string) {
	b.columns.pushString(col...)
}

func (b *SelectBuilder) From(sql string) {
	b.from = sql
}

func (b *SelectBuilder) Where(f func(b *WhereBuilder)) {
	b.where.ops.sep = " and "
	f(&b.where)
}

func (b *SelectBuilder) OrderBy(col string, direction string) {
	p := col
	if direction != "" {
		p += " " + direction
	}
	b.order.push(p)
}

func (b *SelectBuilder) build() (string, []interface{}) {
	if !b.columns.empty() {
		b.push(&b.columns)
	}
	if b.from != "" {
		b.push("from")
		b.push(b.from)
	}
	if !b.where.ops.empty() {
		b.push("where")
		b.push(&b.where)
	}
	if !b.order.empty() {
		b.push("order by")
		b.push(&b.order)
	}
	return b.builder.build()
}
