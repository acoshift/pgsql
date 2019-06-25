package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	q, args := pgstmt.Delete(func(b *pgstmt.DeleteBuilder) {
		b.From("users")
		b.Where(func(b *pgstmt.WhereBuilder) {
			b.Eq("username", "test")
			b.Or(func(b *pgstmt.WhereBuilder) {
				b.Gt("age", 20)
			})
		})
		b.Returning("id", "name")
	})

	assert.Equal(t,
		"delete from users where username = $1 or (age > $2) returning id, name",
		q,
	)
	assert.EqualValues(t,
		[]interface{}{"test", 20},
		args,
	)
}
