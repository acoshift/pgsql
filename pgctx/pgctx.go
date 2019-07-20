package pgctx

import (
	"context"
	"database/sql"
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
}

// NewContext creates new context
func NewContext(ctx context.Context, db DB) context.Context {
	ctx = context.WithValue(ctx, ctxKeyDB{}, db)
	ctx = context.WithValue(ctx, ctxKeyQueryer{}, db)
	return ctx
}

// Middleware injects db into request's context
func Middleware(db DB) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(NewContext(r.Context(), db))
			h.ServeHTTP(w, r)
		})
	}
}

// RunInTx starts sql tx if not started
func RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	// already in tx, do nothing
	if _, ok := ctx.Value(ctxKeyQueryer{}).(*sql.Tx); ok {
		return f(ctx)
	}

	db := ctx.Value(ctxKeyDB{}).(pgsql.BeginTxer)
	var cm *onCommitted
	abort := false
	err := pgsql.RunInTxContext(ctx, db, nil, func(tx *sql.Tx) error {
		cm = &onCommitted{} // reset when retry
		ctx := context.WithValue(ctx, ctxKeyQueryer{}, tx)
		ctx = context.WithValue(ctx, ctxKeyCommitted{}, cm)
		err := f(ctx)
		if err == pgsql.ErrAbortTx {
			abort = true
		}
		return err
	})
	if err != nil {
		return err
	}
	if !abort && cm != nil {
		for _, f := range cm.f {
			f(ctx)
		}
	}
	return nil
}

// Committed calls f after committed
func Committed(ctx context.Context, f func(ctx context.Context)) {
	// check is in tx ?
	if _, ok := ctx.Value(ctxKeyQueryer{}).(*sql.Tx); !ok {
		panic("pgsql: not in tx")
	}

	if f == nil {
		return
	}

	p := ctx.Value(ctxKeyCommitted{}).(*onCommitted)
	p.f = append(p.f, f)
}

type (
	ctxKeyDB        struct{}
	ctxKeyQueryer   struct{}
	ctxKeyCommitted struct{}
)

type onCommitted struct {
	f []func(ctx context.Context)
}

func q(ctx context.Context) Queryer {
	return ctx.Value(ctxKeyQueryer{}).(Queryer)
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
