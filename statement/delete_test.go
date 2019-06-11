package statement_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/statement"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	q, args := statement.Delete(func(b *statement.DeleteBuilder) {
		b.From("users")
		b.Where(func(b *statement.WhereBuilder) {
			b.Eq("username", "test")
			b.Or(func(b *statement.WhereBuilder) {
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
