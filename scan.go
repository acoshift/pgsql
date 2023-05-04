package pgsql

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/lib/pq"
)

// Scan wraps scanner with custom scanner
func Scan(scan Scanner) Scanner {
	return func(dest ...any) error {
		for i, d := range dest {
			// skip known type
			switch d.(type) {
			case sql.Scanner,
				*time.Time,
				*[]byte,
				*int, *int8, *int16, *int32, *int64,
				*uint, *uint8, *uint16, *uint32, *uint64,
				*bool,
				*float32, *float64,
				*sql.RawBytes,
				*sql.Rows:
				continue
			}

			dt := reflect.TypeOf(d)
			if dt.Kind() == reflect.Ptr {
				dt = dt.Elem()
			}
			switch dt.Kind() {
			case reflect.Slice:
				dest[i] = pq.Array(d)
			case reflect.Struct:
				dest[i] = JSON(d)
			}
		}
		return scan(dest...)
	}
}

type Row struct {
	*sql.Row
}

func (r *Row) Scan(dest ...any) error {
	return Scan(r.Row.Scan)(dest...)
}

type Rows struct {
	*sql.Rows
}

func (r *Rows) Scan(dest ...any) error {
	return Scan(r.Rows.Scan)(dest...)
}
