package pgstmt

import (
	"context"
	"database/sql"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
)

type Result struct {
	query string
	args  []any
}

func newResult(query string, args []any) *Result {
	return &Result{query, args}
}

func (r *Result) SQL() (query string, args []any) {
	return r.query, r.args
}

func (r *Result) QueryRow(f func(string, ...any) *sql.Row) *pgsql.Row {
	return &pgsql.Row{f(r.query, r.args...)}
}

func (r *Result) Query(f func(string, ...any) (*sql.Rows, error)) (*pgsql.Rows, error) {
	rows, err := f(r.query, r.args...)
	if err != nil {
		return nil, err
	}
	return &pgsql.Rows{rows}, nil
}

func (r *Result) Exec(f func(string, ...any) (sql.Result, error)) (sql.Result, error) {
	return f(r.query, r.args...)
}

func (r *Result) QueryRowContext(ctx context.Context, f func(context.Context, string, ...any) *sql.Row) *pgsql.Row {
	return &pgsql.Row{f(ctx, r.query, r.args...)}
}

func (r *Result) QueryContext(ctx context.Context, f func(context.Context, string, ...any) (*sql.Rows, error)) (*pgsql.Rows, error) {
	rows, err := f(ctx, r.query, r.args...)
	if err != nil {
		return nil, err
	}
	return &pgsql.Rows{rows}, nil
}

func (r *Result) ExecContext(ctx context.Context, f func(context.Context, string, ...any) (sql.Result, error)) (sql.Result, error) {
	return f(ctx, r.query, r.args...)
}

func (r *Result) QueryRowWith(ctx context.Context) *pgsql.Row {
	return pgctx.QueryRow(ctx, r.query, r.args...)
}

func (r *Result) QueryWith(ctx context.Context) (*pgsql.Rows, error) {
	return pgctx.Query(ctx, r.query, r.args...)
}

func (r *Result) ExecWith(ctx context.Context) (sql.Result, error) {
	return pgctx.Exec(ctx, r.query, r.args...)
}

func (r *Result) IterWith(ctx context.Context, iter pgsql.Iterator) error {
	return pgctx.Iter(ctx, iter, r.query, r.args...)
}
