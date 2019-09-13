package pgmodel

import (
	"github.com/acoshift/pgsql/pgstmt"
)

type Filter interface {
	apply(b pgstmt.SelectStatement)
}

type filterFunc func(b pgstmt.SelectStatement)

func (f filterFunc) apply(b pgstmt.SelectStatement) { f(b) }

func One(field string, eqValue interface{}) Filter {
	return Where(func(b pgstmt.Cond) {
		b.Eq(field, eqValue)
	})
}

func Where(f func(b pgstmt.Cond)) Filter {
	return filterFunc(func(b pgstmt.SelectStatement) {
		b.Where(f)
	})
}

func Having(f func(b pgstmt.Cond)) Filter {
	return filterFunc(func(b pgstmt.SelectStatement) {
		b.Having(f)
	})
}

func OrderBy(col string) Filter {
	return filterFunc(func(b pgstmt.SelectStatement) {
		b.OrderBy(col)
	})
}

func Limit(n int64) Filter {
	return filterFunc(func(b pgstmt.SelectStatement) {
		b.Limit(n)
	})
}

func Offset(n int64) Filter {
	return filterFunc(func(b pgstmt.SelectStatement) {
		b.Offset(n)
	})
}
