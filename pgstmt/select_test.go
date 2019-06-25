package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	t.Run("only select", func(t *testing.T) {
		q, args := pgstmt.Select(func(b *pgstmt.SelectBuilder) {
			b.Columns("1")
		}).SQL()

		assert.Equal(t,
			"select 1",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("select from", func(t *testing.T) {
		q, args := pgstmt.Select(func(b *pgstmt.SelectBuilder) {
			b.Columns("id", "name")
			b.From("users")
		}).SQL()

		assert.Equal(t,
			"select id, name from users",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("select from where order", func(t *testing.T) {
		q, args := pgstmt.Select(func(b *pgstmt.SelectBuilder) {
			b.Columns("id", "name")
			b.From("users")
			b.Where(func(b *pgstmt.WhereBuilder) {
				b.Eq("id", 1)
			})
			b.OrderBy("created_at", "asc")
			b.OrderBy("id", "desc")
		}).SQL()

		assert.Equal(t,
			"select id, name from users where ((id = $1)) order by created_at asc, id desc",
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
