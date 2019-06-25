package statement_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/statement"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	t.Run("only select", func(t *testing.T) {
		q, args := statement.Select(func(b *statement.SelectBuilder) {
			b.Columns("1")
		})

		assert.Equal(t,
			"select 1",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("select from", func(t *testing.T) {
		q, args := statement.Select(func(b *statement.SelectBuilder) {
			b.Columns("id", "name")
			b.From("users")
		})

		assert.Equal(t,
			"select id, name from users",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("select from where", func(t *testing.T) {
		q, args := statement.Select(func(b *statement.SelectBuilder) {
			b.Columns("id", "name")
			b.From("users")
			b.Where(func(b *statement.WhereBuilder) {
				b.Eq("id", 1)
			})
		})

		assert.Equal(t,
			"select id, name from users where (id = $1)",
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				1,
			},
			args,
		)
	})
}
