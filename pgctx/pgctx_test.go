package pgctx_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
)

func newCtx(t *testing.T) (context.Context, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	return pgctx.NewContext(context.Background(), db), mock
}

func TestNewContext(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		newCtx(t)
	})
}

type testKey1 struct{}

func TestNewKeyContext(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		ctx := pgctx.NewKeyContext(context.Background(), testKey1{}, db)
		assert.NotNil(t, ctx)
	})
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	assert.NoError(t, err)

	called := false
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	pgctx.Middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		ctx := r.Context()
		assert.NotPanics(t, func() {
			pgctx.QueryRow(ctx, "select 1")
		})
		assert.NotPanics(t, func() {
			pgctx.Query(ctx, "select 1")
		})
		assert.NotPanics(t, func() {
			pgctx.Exec(ctx, "select 1")
		})
	})).ServeHTTP(w, r)
	assert.True(t, called)
}

func TestKeyMiddleware(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	assert.NoError(t, err)

	called := false
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	pgctx.KeyMiddleware(testKey1{}, db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		ctx := r.Context()
		assert.NotPanics(t, func() {
			pgctx.QueryRow(pgctx.With(ctx, testKey1{}), "select 1")
		})
		assert.NotPanics(t, func() {
			pgctx.Query(pgctx.With(ctx, testKey1{}), "select 1")
		})
		assert.NotPanics(t, func() {
			pgctx.Exec(pgctx.With(ctx, testKey1{}), "select 1")
		})
		assert.Panics(t, func() {
			pgctx.QueryRow(ctx, "select 1")
		})
	})).ServeHTTP(w, r)
	assert.True(t, called)
}

func TestRunInTx(t *testing.T) {
	t.Parallel()

	t.Run("Committed", func(t *testing.T) {
		ctx, mock := newCtx(t)

		called := false
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := pgctx.RunInTx(ctx, func(ctx context.Context) error {
			called = true
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Rollback with error", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectRollback()
		var retErr = fmt.Errorf("error")
		err := pgctx.RunInTx(ctx, func(ctx context.Context) error {
			return retErr
		})
		assert.Error(t, err)
		assert.Equal(t, retErr, err)
	})

	t.Run("Abort Tx", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectCommit()
		err := pgctx.RunInTx(ctx, func(ctx context.Context) error {
			return pgsql.ErrAbortTx
		})
		assert.NoError(t, err)
	})

	t.Run("Nested Tx", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectCommit()
		err := pgctx.RunInTx(ctx, func(ctx context.Context) error {
			return pgctx.RunInTx(ctx, func(ctx context.Context) error {
				return nil
			})
		})
		assert.NoError(t, err)
	})
}

func TestCommitted(t *testing.T) {
	t.Parallel()

	t.Run("Outside Tx", func(t *testing.T) {
		ctx, _ := newCtx(t)
		var called bool
		pgctx.Committed(ctx, func(ctx context.Context) {
			called = true
		})
		assert.True(t, called)
	})

	t.Run("Nil func", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectCommit()
		pgctx.RunInTx(ctx, func(ctx context.Context) error {
			pgctx.Committed(ctx, nil)
			return nil
		})
	})

	t.Run("Committed", func(t *testing.T) {
		ctx, mock := newCtx(t)

		called := false
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := pgctx.RunInTx(ctx, func(ctx context.Context) error {
			pgctx.Committed(ctx, func(ctx context.Context) {
				called = true
			})
			return nil
		})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Rollback", func(t *testing.T) {
		ctx, mock := newCtx(t)

		mock.ExpectBegin()
		mock.ExpectRollback()
		err := pgctx.RunInTx(ctx, func(ctx context.Context) error {
			pgctx.Committed(ctx, func(ctx context.Context) {
				assert.Fail(t, "should not be called")
			})
			return pgsql.ErrAbortTx
		})
		assert.NoError(t, err)
	})
}
