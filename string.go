package pgsql

import (
	"database/sql"
	"database/sql/driver"
)

// NullString scans null into empty string and convert empty string into sql null
func NullString(s *string) interface {
	driver.Valuer
	sql.Scanner
} {
	return Null(s)
}
