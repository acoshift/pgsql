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

type parenGroup struct {
	group
}
