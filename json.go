package pgsql

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON wraps value with scanner and valuer
func JSON(value interface{}) interface {
	driver.Valuer
	sql.Scanner
} {
	return &jsonValue{value}
}

type jsonValue struct {
	value interface{}
}

func (v *jsonValue) Scan(src interface{}) error {
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
