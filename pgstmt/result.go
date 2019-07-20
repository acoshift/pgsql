package pgstmt

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/acoshift/pgsql/pgctx"
)

type Result struct {
	query string
	args  []interface{}
}

func newResult(query string, args []interface{}) *Result {
	return &Result{query, args}
}

func (r *Result) SQL() (query string, args interface{}) {
	return r.query, r.args
}

func (r *Result) QueryRow(f func(string, ...interface{}) *sql.Row) *sql.Row {
	return f(r.query, r.args...)
}

func (r *Result) Query(f func(string, ...interface{}) (*sql.Rows, error)) (*sql.Rows, error) {
	return f(r.query, r.args...)
}

func (r *Result) Exec(f func(string, ...interface{}) (sql.Result, error)) (sql.Result, error) {
	return f(r.query, r.args...)
}

func (r *Result) QueryRowContext(ctx context.Context, f func(context.Context, string, ...interface{}) *sql.Row) *sql.Row {
	return f(ctx, r.query, r.args...)
}

func (r *Result) QueryContext(ctx context.Context, f func(context.Context, string, ...interface{}) (*sql.Rows, error)) (*sql.Rows, error) {
	return f(ctx, r.query, r.args...)
}

func (r *Result) ExecContext(ctx context.Context, f func(context.Context, string, ...interface{}) (sql.Result, error)) (sql.Result, error) {
	return f(ctx, r.query, r.args...)
}

func (r *Result) QueryRowWith(ctx context.Context) *sql.Row {
	return pgctx.QueryRow(ctx, r.query, r.args...)
}

func (r *Result) QueryWith(ctx context.Context) (*sql.Rows, error) {
	return pgctx.Query(ctx, r.query, r.args...)
}

func (r *Result) ExecWith(ctx context.Context) (sql.Result, error) {
	return pgctx.Exec(ctx, r.query, r.args...)
}

func (r *Result) QueryAllWith(ctx context.Context, to interface{}, each func(scan func(dest ...interface{}) error) (interface{}, error)) error {
	refTo := reflect.ValueOf(to).Elem()

	rows, err := r.QueryWith(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		x, err := each(rows.Scan)
		if err != nil {
			return err
		}
		refTo = reflect.Append(refTo, reflect.ValueOf(x))
	}

	reflect.ValueOf(to).Elem().Set(refTo)

	return rows.Err()
}
