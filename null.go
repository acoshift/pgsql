package pgsql

import (
	"database/sql"
	"database/sql/driver"
	_ "unsafe"
)

func Null[T comparable](v *T) interface {
	driver.Valuer
	sql.Scanner
} {
	return &null[T]{v}
}

type null[T comparable] struct {
	value *T
}

func (s *null[T]) Scan(src interface{}) error {
	*s.value = *(new(T))
	if src == nil {
		return nil
	}

	return convertAssign(s.value, src)
}

func (s null[T]) Value() (driver.Value, error) {
	if s.value == nil || isZero(*s.value) || *s.value == *(new(T)) {
		return nil, nil
	}
	return *s.value, nil
}

type zeroer interface {
	IsZero() bool
}

func isZero(v any) bool {
	if z, ok := v.(zeroer); ok {
		return z.IsZero()
	}
	return false
}

//go:linkname convertAssign database/sql.convertAssign
func convertAssign(dst, src any) error
