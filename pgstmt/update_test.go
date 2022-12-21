package pgstmt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("update", func(t *testing.T) {
		q, args := pgstmt.Update(func(b pgstmt.UpdateStatement) {
			b.Table("users")
			b.Set("name").To("test")
			b.Set("email", "address", "updated_at").To("test@localhost", "123", pgstmt.Raw("now()"))
			b.Set("age").ToRaw(1)
			b.Where(func(b pgstmt.Cond) {
				b.Eq("id", 5)
			})
			b.Returning("id", "name")
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				update users
				set name = $1,
					(email, address, updated_at) = row($2, $3, now()),
					age = 1
				where (id = $4)
				returning id, name
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				"test",
				"test@localhost", "123",
				5,
			},
			args,
		)
	})

	t.Run("update set select", func(t *testing.T) {
		q, args := pgstmt.Update(func(b pgstmt.UpdateStatement) {
			b.Table("users")
			b.Set("name", "age", "updated_at").Select(func(b pgstmt.SelectStatement) {
				b.Columns("name", "age", "now()")
				b.From("users")
				b.Where(func(b pgstmt.Cond) {
					b.Eq("id", 6)
				})
			})
			b.Set("updated_count").ToRaw("updated_count + 1")
			b.Set("email", "address").To("test@localhost", "123")
			b.Where(func(b pgstmt.Cond) {
				b.Eq("id", 5)
			})
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				update users
				set (name, age, updated_at) = (select name, age, now()
											   from users
											   where (id = $1)),
					updated_count = updated_count + 1,
					(email, address) = row($2, $3)
				where (id = $4)
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				6,
				"test@localhost", "123",
				5,
			},
			args,
		)
	})

	t.Run("update from join", func(t *testing.T) {
		q, args := pgstmt.Update(func(b pgstmt.UpdateStatement) {
			b.Table("users")
			b.Set("name").ToRaw("p.name")
			b.Set("address").ToRaw("p.address")
			b.Set("updated_at").ToRaw("now()")
			b.Set("date").To(pgstmt.NotArg(time.Date(2022, 1, 2, 3, 4, 5, 6, time.UTC)))
			b.From("users")
			b.InnerJoin("profiles p").Using("email")
			b.Where(func(b pgstmt.Cond) {
				b.Eq("users.id", 2)
			})
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				update users
				set name = p.name,
					address = p.address,
					updated_at = now(),
					date = '2022-01-02 03:04:05.000000006Z'
				from users
				inner join profiles p using (email)
				where (users.id = $1)
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				2,
			},
			args,
		)
	})
}
