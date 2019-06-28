package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	t.Run("insert", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "name", "created_at")
			b.Value("tester1", "Tester 1", pgstmt.Default)
			b.Value("tester2", "Tester 2", "now()")
			b.OnConflict("username").DoNothing()
			b.Returning("id", "name")
		}).SQL()

		assert.Equal(t,
			"insert into users (username, name, created_at) values ($1, $2, default), ($3, $4, $5) on conflict (username) do nothing returning id, name",
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				"tester1", "Tester 1",
				"tester2", "Tester 2", "now()",
			},
			args,
		)
	})
}
