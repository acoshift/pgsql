package pgsql

import (
	"context"
	"database/sql"
)

type Scanner func(dest ...interface{}) error

type Iterator func(scan Scanner) error

// QueryContext interface
type QueryContext interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

func Iter(q interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}, iter Iterator, query string, args ...interface{}) error {
	return IterContext(context.Background(), q, iter, query, args...)
}

func IterContext(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}, iter Iterator, query string, args ...interface{}) error {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := iter(rows.Scan)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}
