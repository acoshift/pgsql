package statement

import (
	"strconv"
	"strings"
)

type builder struct {
	q []interface{}
}

func (b *builder) push(q ...interface{}) {
	b.q = append(b.q, q...)
}

func (b *builder) build() (string, []interface{}) {
	var args []interface{}
	var i int

	var f func(p []interface{}, sep string) string
	f = func(p []interface{}, sep string) string {
		var q []string
		for _, x := range p {
			switch x := x.(type) {
			case string:
				q = append(q, x)
			case argWrapper:
				i++
				q = append(q, "$"+strconv.Itoa(i))
				args = append(args, x.value)
			case group:
				if !x.empty() {
					q = append(q, f(x.q, x.getSep()))
				}
			case parenGroup:
				if !x.empty() {
					q = append(q, "("+f(x.q, x.getSep())+")")
				}
			}
		}
		return strings.Join(q, sep)
	}
	query := f(b.q, " ")
	return query, args
}

func arg(v interface{}) interface{} {
	return argWrapper{v}
}

type argWrapper struct {
	value interface{}
}

type group struct {
	q   []interface{}
	sep string
}

func (b group) getSep() string {
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
