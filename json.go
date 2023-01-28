package pgsql

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON wraps value with scanner and valuer
func JSON(value any) interface {
	driver.Valuer
	sql.Scanner
} {
	return &jsonValue{value}
}

type jsonValue struct {
	value any
}

func (v *jsonValue) Scan(src any) error {
	if src == nil {
		return nil
	}

	var b []byte
	switch p := src.(type) {
	case []byte:
		b = p
	case string:
		b = []byte(p)
	default:
		return fmt.Errorf("pgsql: JSON not support scan source")
	}
	return json.Unmarshal(b, v.value)
}

func (v jsonValue) Value() (driver.Value, error) {
	return json.Marshal(v.value)
}
