package pgstmt

import (
	"fmt"
	"strconv"
	"strings"
)

type buffer struct {
	q []interface{}
}

func (b *buffer) push(q ...interface{}) {
	b.q = append(b.q, q...)
}

func (b *buffer) pushFirst(q ...interface{}) {
	b.q = append(q, b.q...)
}

func (b *buffer) empty() bool {
	return len(b.q) == 0
}

func (b *buffer) build() []interface{} {
	return b.q
}

type builder interface {
	build() []interface{}
}

func build(b *buffer) (string, []interface{}) {
	var args []interface{}
	var i int

	var f func(p []interface{}, sep string) string
	f = func(p []interface{}, sep string) string {
		var q []string
		for _, x := range p {
			switch x := x.(type) {
			default:
				q = append(q, convertToString(x))
			case builder:
				q = append(q, f(x.build(), " "))
			case arg:
				i++
				q = append(q, "$"+strconv.Itoa(i))
				args = append(args, x.value)
			case *group:
				if !x.empty() {
					q = append(q, f(x.q, x.getSep()))
				}
			case *parenGroup:
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

func convertToString(x interface{}) string {
	switch x := x.(type) {
	default:
		return fmt.Sprint(x)
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	case notArg:
		return convertToString(x.value)
	}
}
