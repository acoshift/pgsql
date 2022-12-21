package pgstmt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type buffer struct {
	q []interface{}
}

func (b *buffer) push(q ...interface{}) {
	b.q = append(b.q, q...)
}

func (b *buffer) pushFront(q ...interface{}) {
	b.q = append(q, b.q...)
}

func (b *buffer) popFront() interface{} {
	if b.empty() {
		return nil
	}
	p := b.q[0]
	b.q = b.q[1:]
	return p
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
				q = append(q, convertToString(x, false))
			case builder:
				q = append(q, f(x.build(), " "))
			case arg:
				i++
				q = append(q, "$"+strconv.Itoa(i))
				args = append(args, x.value)
			case _any:
				i++
				q = append(q, fmt.Sprintf("any($%d)", i))
				args = append(args, x.value)
			case all:
				i++
				q = append(q, fmt.Sprintf("all($%d)", i))
				args = append(args, x.value)
			case *group:
				if !x.empty() {
					q = append(q, f(x.q, x.getSep()))
				}
			case *parenGroup:
				if !x.empty() {
					q = append(q, x.prefix+"("+f(x.q, x.getSep())+")")
				}
			}
		}
		return strings.Join(q, sep)
	}
	query := f(b.q, " ")
	return query, args
}

func convertToString(x interface{}, quoteStr bool) string {
	switch x := x.(type) {
	default:
		return fmt.Sprint(x)
	case string:
		if quoteStr {
			return pq.QuoteLiteral(x)
		}
		return x
	case int:
		return strconv.Itoa(x)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	case time.Time:
		return convertToString(string(pq.FormatTimestamp(x)), true)
	case notArg:
		return convertToString(x.value, true)
	case raw:
		return fmt.Sprint(x.value)
	case defaultValue:
		return "default"
	}
}
