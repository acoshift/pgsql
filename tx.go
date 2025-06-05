package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// ErrAbortTx rollbacks transaction and return nil error
var ErrAbortTx = errors.New("pgsql: abort tx")

// BeginTxer type
type BeginTxer interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// BackoffDelayFunc is a function type that defines the delay for backoff
type BackoffDelayFunc func(attempt int) time.Duration

// TxOptions is the transaction options
type TxOptions struct {
	sql.TxOptions
	MaxAttempts      int
	BackoffDelayFunc BackoffDelayFunc
}

const (
	defaultMaxAttempts = 10
)

// RunInTx runs fn inside retryable transaction.
//
// see RunInTxContext for more info.
func RunInTx(db BeginTxer, opts *TxOptions, fn func(*sql.Tx) error) error {
	return RunInTxContext(context.Background(), db, opts, fn)
}

// RunInTxContext runs fn inside retryable transaction with context.
// It use Serializable isolation level if tx options isolation is setted to sql.LevelDefault.
//
// RunInTxContext DO NOT handle panic.
// But when panic, it will rollback the transaction.
func RunInTxContext(ctx context.Context, db BeginTxer, opts *TxOptions, fn func(*sql.Tx) error) error {
	option := TxOptions{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
		MaxAttempts: defaultMaxAttempts,
	}

	if opts != nil {
		if opts.MaxAttempts > 0 {
			option.MaxAttempts = opts.MaxAttempts
		}
		option.TxOptions = opts.TxOptions

		// override default isolation level to serializable
		if opts.Isolation == sql.LevelDefault {
			option.Isolation = sql.LevelSerializable
		}

		option.BackoffDelayFunc = opts.BackoffDelayFunc
	}

	var backoffTimer *backoffTimer
	if option.BackoffDelayFunc != nil {
		backoffTimer = newBackoffTimer(option.BackoffDelayFunc)
		defer backoffTimer.Stop()
	}

	f := func() error {
		tx, err := db.BeginTx(ctx, &option.TxOptions)
		if err != nil {
			return err
		}
		// use defer to also rollback when panic
		defer tx.Rollback()

		err = fn(tx)
		if err != nil {
			return err
		}
		return tx.Commit()
	}

	var err error
	for i := 0; i < option.MaxAttempts; i++ {
		err = f()
		if err == nil || errors.Is(err, ErrAbortTx) {
			return nil
		}
		if !IsSerializationFailure(err) {
			return err
		}

		if backoffTimer != nil && i < option.MaxAttempts-1 {
			if err = backoffTimer.Wait(ctx, i); err != nil {
				return err
			}
		}
	}

	return err
}

type backoffTimer struct {
	timer            *time.Timer
	backOffDelayFunc BackoffDelayFunc
}

func newBackoffTimer(backoffDelayFunc BackoffDelayFunc) *backoffTimer {
	return &backoffTimer{
		timer:            time.NewTimer(0),
		backOffDelayFunc: backoffDelayFunc,
	}
}

func (b *backoffTimer) Wait(ctx context.Context, attempt int) error {
	delay := b.backOffDelayFunc(attempt)
	if delay <= 0 {
		return nil
	}

	b.timer.Reset(delay)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.timer.C:
		return nil
	}
}

func (b *backoffTimer) Stop() bool {
	return b.timer.Stop()
}
