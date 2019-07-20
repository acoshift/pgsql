package pgstmt

type group struct {
	q   []interface{}
	sep string
}

func (b *group) getSep() string {
	if b.sep == "" {
		return ", "
	}
	return b.sep
}

func (b *group) empty() bool {
	return len(b.q) == 0
}

func (b *group) push(q ...interface{}) {
	b.q = append(b.q, q...)
}

func (b *group) pushString(q ...string) {
	for _, x := range q {
		b.q = append(b.q, x)
	}
}

func withGroup(sep string, q ...interface{}) interface{} {
	var g group
	g.sep = sep
	g.push(q...)
	return &g
}

type parenGroup struct {
	group
	prefix string
}

func paren(q ...interface{}) interface{} {
	var p parenGroup
	p.push(q...)
	return &p
}

func parenString(q ...string) interface{} {
	var p parenGroup
	p.pushString(q...)
	return &p
}

func withParen(sep string, q ...interface{}) interface{} {
	var p parenGroup
	p.sep = sep
	p.push(q...)
	return &p
}
