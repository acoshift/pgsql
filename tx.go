package pgsql

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// TxOptions is the transaction options
type TxOptions struct {
	sql.TxOptions
	MaxAttempts int
}

const (
	defaultMaxAttempts = 10
)

// Tx interface
type Tx interface {
	Commit() error
	Rollback() error
}

// RunInTx runs fn inside retryable transaction
func RunInTx(db *sql.DB, opts *TxOptions, fn func(*sql.Tx) error) error {
	return RunInTxContext(context.Background(), db, opts, fn)
}

// RunInTxContext runs fn inside retryable transaction with context
func RunInTxContext(ctx context.Context, db *sql.DB, opts *TxOptions, fn func(*sql.Tx) error) (err error) {
	if opts == nil {
		opts = &TxOptions{
			MaxAttempts: defaultMaxAttempts,
		}
	}
	// override default isolation level to serializable
	if opts.Isolation == sql.LevelDefault {
		opts.Isolation = sql.LevelSerializable
	}

	var tx *sql.Tx
	for i := 0; i < opts.MaxAttempts; i++ {
		tx, err = db.BeginTx(ctx, &opts.TxOptions)
		if err != nil {
			return
		}

		err = fn(tx)
		if err == nil {
			if err = tx.Commit(); err == nil {
				return
			}
		}
		pqErr, ok := err.(*pq.Error)
		if retryable := ok && (pqErr.Code == "40001"); !retryable {
			tx.Rollback() // ignore rollback error
			return
		}
	}
	return
}
