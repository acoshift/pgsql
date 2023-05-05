package pgsql_test

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql"
)

func TestNull_Value(t *testing.T) {
	t.Parallel()

	t.Run("Int64 valid", func(t *testing.T) {
		x := 1
		v, err := pgsql.Null(&x).Value()
		assert.NoError(t, err)
		assert.Equal(t, x, v)
	})

	t.Run("Int64 zero", func(t *testing.T) {
		x := 0
		v, err := pgsql.Null(&x).Value()
		assert.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("Valuer valid", func(t *testing.T) {
		x := testValuer{1}
		v, err := pgsql.Null(&x).Value()
		assert.NoError(t, err)
		assert.Equal(t, x.x, v)
	})

	t.Run("Valuer zero", func(t *testing.T) {
		x := testValuer{0}
		v, err := pgsql.Null(&x).Value()
		assert.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("Valuer nil", func(t *testing.T) {
		var x *testValuer
		v, err := pgsql.Null(x).Value()
		assert.NoError(t, err)
		assert.Nil(t, v)
	})
}

type testValuer struct {
	x int64
}

func (v testValuer) Value() (driver.Value, error) {
	return v.x, nil
}

func (v testValuer) IsZero() bool {
	return v.x == 0
}
