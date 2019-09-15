package pgmodel

import (
	"context"

	"github.com/acoshift/pgsql/pgstmt"
)

type Filter interface {
	Apply(ctx context.Context, b pgstmt.SelectStatement)
}

type FilterFunc func(ctx context.Context, b pgstmt.SelectStatement)

func (f FilterFunc) Apply(ctx context.Context, b pgstmt.SelectStatement) { f(ctx, b) }

func Equal(field string, value interface{}) Filter {
	return Where(func(b pgstmt.Cond) {
		b.Eq(field, value)
	})
}

func Where(f func(b pgstmt.Cond)) Filter {
	return FilterFunc(func(_ context.Context, b pgstmt.SelectStatement) {
		b.Where(f)
	})
}

func Having(f func(b pgstmt.Cond)) Filter {
	return FilterFunc(func(_ context.Context, b pgstmt.SelectStatement) {
		b.Having(f)
	})
}

func OrderBy(col string) Filter {
	return FilterFunc(func(_ context.Context, b pgstmt.SelectStatement) {
		b.OrderBy(col)
	})
}

func Limit(n int64) Filter {
	return FilterFunc(func(_ context.Context, b pgstmt.SelectStatement) {
		b.Limit(n)
	})
}

func Offset(n int64) Filter {
	return FilterFunc(func(_ context.Context, b pgstmt.SelectStatement) {
		b.Offset(n)
	})
}
