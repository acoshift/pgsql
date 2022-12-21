package pgctx

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/acoshift/pgsql"
)

type DB interface {
	Queryer
	pgsql.BeginTxer
}

// Queryer interface
type Queryer interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

func NewKeyContext(ctx context.Context, key any, db DB) context.Context {
	return context.WithValue(ctx, ctxKeyDB{key}, db)
}

// NewContext creates new context
func NewContext(ctx context.Context, db DB) context.Context {
	return NewKeyContext(ctx, nil, db)
}

func KeyMiddleware(key any, db DB) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(NewKeyContext(r.Context(), key, db))
			h.ServeHTTP(w, r)
		})
	}
}

// Middleware injects db into request's context
func Middleware(db DB) func(h http.Handler) http.Handler {
	return KeyMiddleware(nil, db)
}

// With creates new empty key context with db from keyed context
func With(ctx context.Context, key any) context.Context {
	db := ctx.Value(ctxKeyDB{key})
	return context.WithValue(ctx, ctxKeyDB{}, db)
}

func GetDB(ctx context.Context) DB {
	return ctx.Value(ctxKeyDB{}).(DB)
}

func GetDBKey(ctx context.Context, key any) DB {
	return ctx.Value(ctxKeyDB{key}).(DB)
}

type wrapTx struct {
	*sql.Tx
	onCommitted []func(ctx context.Context)
}

var _ Queryer = &wrapTx{}

// RunInTxOptions starts sql tx if not started
func RunInTxOptions(ctx context.Context, opt *pgsql.TxOptions, f func(ctx context.Context) error) error {
	if IsInTx(ctx) {
		return f(ctx)
	}

	db := ctx.Value(ctxKeyDB{}).(pgsql.BeginTxer)
	var pTx wrapTx
	abort := false
	err := pgsql.RunInTxContext(ctx, db, opt, func(tx *sql.Tx) error {
		pTx = wrapTx{Tx: tx}
		ctx := context.WithValue(ctx, ctxKeyQueryer{}, &pTx)
		err := f(ctx)
		if errors.Is(err, pgsql.ErrAbortTx) {
			abort = true
		}
		return err
	})
	if err != nil {
		return err
	}
	if !abort && len(pTx.onCommitted) > 0 {
		for _, f := range pTx.onCommitted {
			f(ctx)
		}
	}
	return nil
}

// RunInTx calls RunInTxOptions with default options
func RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	return RunInTxOptions(ctx, nil, f)
}

// IsInTx checks is context inside RunInTx
func IsInTx(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKeyQueryer{}).(*wrapTx)
	return ok
}

// Committed calls f after committed or immediate if not in tx
func Committed(ctx context.Context, f func(ctx context.Context)) {
	if f == nil {
		return
	}

	if !IsInTx(ctx) {
		f(ctx)
		return
	}

	pTx := ctx.Value(ctxKeyQueryer{}).(*wrapTx)
	pTx.onCommitted = append(pTx.onCommitted, f)
}

type (
	ctxKeyDB struct {
		key any
	}
	ctxKeyQueryer struct{}
)

func q(ctx context.Context) Queryer {
	if q, ok := ctx.Value(ctxKeyQueryer{}).(Queryer); ok {
		return q
	}
	return ctx.Value(ctxKeyDB{}).(Queryer)
}

// QueryRow calls db.QueryRowContext
func QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return q(ctx).QueryRowContext(ctx, query, args...)
}

// Query calls db.QueryContext
func Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return q(ctx).QueryContext(ctx, query, args...)
}

// Exec calls db.ExecContext
func Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return q(ctx).ExecContext(ctx, query, args...)
}

// Iter calls pgsql.IterContext
func Iter(ctx context.Context, iter pgsql.Iterator, query string, args ...interface{}) error {
	return pgsql.IterContext(ctx, q(ctx), iter, query, args...)
}

// Prepare calls db.PrepareContext
func Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return q(ctx).PrepareContext(ctx, query)
}
