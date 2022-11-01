package pgsql_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/acoshift/pgsql"
)

func TestNullInt64(t *testing.T) {
	db := open(t)

	_, err := db.Exec(`
		drop table if exists test_pgsql_null_int64;
		create table test_pgsql_null_int64 (
			id int primary key,
			value int
		);
		insert into test_pgsql_null_int64 (
			id, value
		) values
			(0, 1),
			(1, null);
	`)
	require.NoError(t, err)
	defer db.Exec(`drop table test_pgsql_null_int64`)

	t.Run("Scan", func(t *testing.T) {
		{
			var p int64
			err = db.QueryRow(`select value from test_pgsql_null_int64 where id = 0`).Scan(pgsql.NullInt64(&p))
			require.NoError(t, err)
			assert.Equal(t, int64(1), p)
		}

		{
			var p int64
			err = db.QueryRow(`select value from test_pgsql_null_int64 where id = 1`).Scan(pgsql.NullInt64(&p))
			require.NoError(t, err)
			assert.Equal(t, int64(0), p)
		}
	})
}
