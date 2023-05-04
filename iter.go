package pgsql

import (
	"context"
	"database/sql"
)

type Scanner func(dest ...any) error

type Iterator func(scan Scanner) error

// QueryContext interface
type QueryContext interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func Iter(q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, iter Iterator, query string, args ...any) error {
	return IterContext(context.Background(), q, iter, query, args...)
}

func IterContext(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, iter Iterator, query string, args ...any) error {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := iter(Scan(rows.Scan))
		if err != nil {
			return err
		}
	}

	return rows.Err()
}
