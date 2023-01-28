package pgsql

import (
	"database/sql"
	"database/sql/driver"
)

// NullInt64 scans null into zero int64 and convert zero into sql null
func NullInt64(i *int64) interface {
	driver.Valuer
	sql.Scanner
} {
	return Null(i)
}
