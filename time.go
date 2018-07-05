package pgsql

import (
	"database/sql/driver"
	"time"
)

// Time is the time.Time but can scan null into empty
type Time struct {
	time.Time
}

// Scan implements Scanner interface
func (t *Time) Scan(src interface{}) error {
	t.Time, _ = src.(time.Time)
	return nil
}

// Value implements Valuer interface
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}
