package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	q, args := pgstmt.Delete(func(b pgstmt.DeleteStatement) {
		b.From("users")
		b.Where(func(b pgstmt.Cond) {
			b.Eq("username", "test")
			b.Eq("is_active", false)
			b.Or(func(b pgstmt.Cond) {
				b.Gt("age", pgstmt.Arg(20))
				b.Le("age", pgstmt.Arg(30))
			})
		})
		b.Returning("id", "name")
	}).SQL()

	assert.Equal(t,
		"delete from users where (username = $1 and is_active = $2) or (age > $3 and age <= $4) returning id, name",
		q,
	)
	assert.EqualValues(t,
		[]any{"test", false, 20, 30},
		args,
	)
}
