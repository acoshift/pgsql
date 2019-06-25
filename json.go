package pgsql

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONObject wraps object with scanner and valuer
func JSONObject(object interface{}) interface {
	driver.Valuer
	sql.Scanner
} {
	return &jsonObject{object}
}

type jsonObject struct {
	Object interface{}
}

func (obj *jsonObject) Scan(src interface{}) error {
	var b []byte
	switch p := src.(type) {
	case []byte:
		b = p
	case string:
		b = []byte(p)
	default:
		return fmt.Errorf("pgsql: JSONObject not support scan source")
	}
	return json.Unmarshal(b, obj.Object)
}

func (obj jsonObject) Value() (driver.Value, error) {
	return json.Marshal(obj.Object)
}
