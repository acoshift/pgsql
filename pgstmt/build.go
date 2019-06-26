package pgstmt

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

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

type extractor interface {
	extract() []interface{}
}

type builder struct {
	q []interface{}
}

func (b *builder) push(q ...interface{}) {
	b.q = append(b.q, q...)
}

func (b *builder) pushFirst(q ...interface{}) {
	b.q = append(q, b.q...)
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
			case extractor:
				q = append(q, f(x.extract(), " "))
			case argWrapper:
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
