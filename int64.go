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
	return &nullInt64{i}
}

type nullInt64 struct {
	value *int64
}

func (i *nullInt64) Scan(src interface{}) error {
	var t sql.NullInt64
	err := t.Scan(src)
	*i.value = t.Int64
	return err
}

func (i nullInt64) Value() (driver.Value, error) {
	if i.value == nil || *i.value == 0 {
		return nil, nil
	}
	return *i.value, nil
}
