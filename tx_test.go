package pgsql_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/acoshift/pgsql"
)

func TestTx(t *testing.T) {
	db := open(t)
	defer db.Close()

	_, err := db.Exec(`
		drop table if exists test_pgsql_tx;
		create table test_pgsql_tx (
			id int primary key,
			value int
		);
		insert into test_pgsql_tx (
			id, value
		) values
			(0, 0),
			(1, 0),
			(2, 0);
	`)
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
	}
	defer db.Exec(`drop table test_pgsql_tx`)
	db.SetMaxOpenConns(30)

	opts := &pgsql.TxOptions{MaxAttempts: 10}

	deposit := func(balance int) error {
		return pgsql.RunInTx(db, opts, func(tx *sql.Tx) error {
			var err error

			// log.Println("deposit", balance)
			var acc0, acc1 int
			err = tx.QueryRow(`select value from test_pgsql_tx where id = 0`).Scan(&acc0)
			if err != nil {
				return err
			}
			err = tx.QueryRow(`select value from test_pgsql_tx where id = 1`).Scan(&acc1)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_pgsql_tx set value = $1 where id = 0`, acc0-balance)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_pgsql_tx set value = $1 where id = 1`, acc1+balance)
			if err != nil {
				return err
			}
			return nil
		})
	}
	withdraw := func(balance int) error {
		return pgsql.RunInTx(db, opts, func(tx *sql.Tx) error {
			var err error

			// log.Println("withdraw", balance)
			var acc0, acc1 int
			err = tx.QueryRow(`select value from test_pgsql_tx where id = 1`).Scan(&acc1)
			if err != nil {
				return err
			}
			if acc1 < balance {
				return fmt.Errorf("not enough balance to withdraw")
			}
			err = tx.QueryRow(`select value from test_pgsql_tx where id = 0`).Scan(&acc0)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_pgsql_tx set value = $1 where id = 0`, acc0+balance)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_pgsql_tx set value = $1 where id = 1`, acc1-balance)
			if err != nil {
				return err
			}
			return nil
		})
	}
	transfer := func(balance int) error {
		return pgsql.RunInTx(db, opts, func(tx *sql.Tx) error {
			var err error

			// log.Println("transfer", balance)
			var acc1, acc2 int
			err = tx.QueryRow(`select value from test_pgsql_tx where id = 1`).Scan(&acc1)
			if err != nil {
				return err
			}
			if acc1 < balance {
				return fmt.Errorf("not enough balance to transfer")
			}
			err = tx.QueryRow(`select value from test_pgsql_tx where id = 2`).Scan(&acc2)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_pgsql_tx set value = $1 where id = 1`, acc1-balance)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`update test_pgsql_tx set value = $1 where id = 2`, acc2+balance)
			if err != nil {
				return err
			}
			return nil
		})
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			var err error
			k := rand.Intn(3)
			if k == 0 {
				err = deposit(rand.Intn(100000))
			} else if k == 1 {
				err = withdraw(rand.Intn(100000))
			} else {
				err = transfer(rand.Intn(100000))
			}
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	var result int
	err = db.QueryRow(`select sum(value) from test_pgsql_tx`).Scan(&result)
	if err != nil {
		t.Fatalf("query result error; %v", err)
	}
	if result != 0 {
		t.Fatalf("expected sum all value to be 0; got %d", result)
	}
}

func TestTxRetryWithBackoff(t *testing.T) {
	t.Parallel()

	t.Run("Backoff when serialization failure occurs", func(t *testing.T) {
		t.Parallel()

		attemptCount := 0
		opts := &pgsql.TxOptions{
			MaxAttempts: 3,
			BackoffDelayFunc: func(attempt int) time.Duration {
				attemptCount++
				return 1
			},
		}

		pgsql.RunInTxContext(context.Background(), sql.OpenDB(&fakeConnector{}), opts, func(*sql.Tx) error {
			return &mockSerializationFailureError{}
		})

		if attemptCount != opts.MaxAttempts-1 {
			t.Fatalf("expected BackoffDelayFunc to be called %d times, got %d", opts.MaxAttempts, attemptCount)
		}
	})

	t.Run("Successful After Multiple Failures", func(t *testing.T) {
		t.Parallel()

		failCount := 0
		maxFailures := 3
		opts := &pgsql.TxOptions{
			MaxAttempts: maxFailures + 1,
			BackoffDelayFunc: func(attempt int) time.Duration {
				return 1
			},
		}

		err := pgsql.RunInTxContext(context.Background(), sql.OpenDB(&fakeConnector{}), opts, func(tx *sql.Tx) error {
			if failCount < maxFailures {
				failCount++
				return &mockSerializationFailureError{}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("expected success after failures, got error: %v", err)
		}
		if failCount != maxFailures {
			t.Fatalf("expected %d failures before success, got %d", maxFailures, failCount)
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		t.Parallel()

		opts := &pgsql.TxOptions{
			MaxAttempts: 3,
			BackoffDelayFunc: func(attempt int) time.Duration {
				return 1
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		err := pgsql.RunInTxContext(ctx, sql.OpenDB(&fakeConnector{}), opts, func(*sql.Tx) error {
			return &mockSerializationFailureError{}
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled error, got %v", err)
		}
	})

	t.Run("Max Attempts Reached", func(t *testing.T) {
		t.Parallel()

		attemptCount := 0
		opts := &pgsql.TxOptions{
			MaxAttempts: 3,
			BackoffDelayFunc: func(attempt int) time.Duration {
				return 1
			},
		}

		err := pgsql.RunInTxContext(context.Background(), sql.OpenDB(&fakeConnector{}), opts, func(*sql.Tx) error {
			attemptCount++
			return &mockSerializationFailureError{}
		})
		if errors.As(err, &mockSerializationFailureError{}) {
			t.Fatal("expected an error when max attempts reached")
		}
		if attemptCount != opts.MaxAttempts {
			t.Fatalf("expected %d attempts, got %d", opts.MaxAttempts, attemptCount)
		}
	})
}

type fakeConnector struct {
	driver.Connector
}

func (c *fakeConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return &fakeConn{}, nil
}

func (c *fakeConnector) Driver() driver.Driver {
	panic("not implemented")
}

type fakeConn struct {
	driver.Conn
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) Begin() (driver.Tx, error) {
	return &fakeTx{}, nil
}

var _ driver.ConnBeginTx = (*fakeConn)(nil)

func (c *fakeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return &fakeTx{}, nil
}

type fakeTx struct {
	driver.Tx
}

func (tx *fakeTx) Commit() error {
	return nil
}

func (tx *fakeTx) Rollback() error {
	return nil
}

type mockSerializationFailureError struct{}

func (e mockSerializationFailureError) Error() string {
	return "mock serialization failure error"
}

func (e mockSerializationFailureError) SQLState() string {
	return "40001" // SQLSTATE code for serialization failure
}
