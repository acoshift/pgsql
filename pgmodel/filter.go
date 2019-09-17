package pgmodel

import (
	"context"

	"github.com/acoshift/pgsql/pgstmt"
)

type Cond interface {
	Where(f func(b pgstmt.Cond))
	Having(f func(b pgstmt.Cond))
	OrderBy(col string) pgstmt.OrderBy
	Limit(n int64)
	Offset(n int64)
}

type Filter interface {
	Apply(ctx context.Context, b Cond)
}

type FilterFunc func(ctx context.Context, b Cond)

func (f FilterFunc) Apply(ctx context.Context, b Cond) { f(ctx, b) }

func Equal(field string, value interface{}) Filter {
	return Where(func(b pgstmt.Cond) {
		b.Eq(field, value)
	})
}

func Where(f func(b pgstmt.Cond)) Filter {
	return FilterFunc(func(_ context.Context, b Cond) {
		b.Where(f)
	})
}

func Having(f func(b pgstmt.Cond)) Filter {
	return FilterFunc(func(_ context.Context, b Cond) {
		b.Having(f)
	})
}

func OrderBy(col string) Filter {
	return FilterFunc(func(_ context.Context, b Cond) {
		b.OrderBy(col)
	})
}

func Limit(n int64) Filter {
	return FilterFunc(func(_ context.Context, b Cond) {
		b.Limit(n)
	})
}

func Offset(n int64) Filter {
	return FilterFunc(func(_ context.Context, b Cond) {
		b.Offset(n)
	})
}

type condUpdateWrapper struct {
	pgstmt.UpdateStatement
}

func (c condUpdateWrapper) Having(f func(b pgstmt.Cond)) {}

func (c condUpdateWrapper) OrderBy(col string) pgstmt.OrderBy { return noopOrderBy{} }

func (c condUpdateWrapper) Limit(n int64) {}

func (c condUpdateWrapper) Offset(n int64) {}

type noopOrderBy struct{}

func (n noopOrderBy) Asc() pgstmt.OrderBy { return n }

func (n noopOrderBy) Desc() pgstmt.OrderBy { return n }

func (n noopOrderBy) NullsFirst() pgstmt.OrderBy { return n }

func (n noopOrderBy) NullsLast() pgstmt.OrderBy { return n }
