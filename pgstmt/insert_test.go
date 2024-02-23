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
			b.OnConflictIndex("username").DoNothing()
			b.Returning("id", "name")
		}).SQL()

		assert.Equal(t,
			"insert into users (username, name, created_at) values ($1, $2, default), ($3, $4, $5) on conflict (username) do nothing returning id, name",
			q,
		)
		assert.EqualValues(t,
			[]any{
				"tester1", "Tester 1",
				"tester2", "Tester 2", "now()",
			},
			args,
		)
	})

	t.Run("insert on conflict do nothing", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "name")
			b.Value("tester1", "Tester 1")
			b.OnConflictDoNothing()
			b.Returning("id")
		}).SQL()

		assert.Equal(t,
			"insert into users (username, name) values ($1, $2) on conflict do nothing returning id",
			q,
		)
		assert.EqualValues(t,
			[]any{
				"tester1", "Tester 1",
			},
			args,
		)
	})

	t.Run("insert select", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("films")
			b.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("tmp_films")
				b.Where(func(b pgstmt.Cond) {
					b.LtRaw("date_prod", "2004-05-07")
				})
			})
		}).SQL()

		assert.Equal(t,
			"insert into films select * from tmp_films where (date_prod < 2004-05-07)",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("insert on conflict partial index do update", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "email")
			b.Value("tester1", "tester1@localhost")
			b.OnConflict(func(b pgstmt.ConflictTarget) {
				b.Index("username")
				b.Where(func(b pgstmt.Cond) {
					b.IsNull("deleted_at")
				})
			}).DoUpdate(func(b pgstmt.UpdateStatement) {
				b.Set("email").ToRaw("excluded.email")
				b.Set("updated_at").ToRaw("now()")
			})
			b.Returning("id")
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				insert into users (username, email)
				values ($1, $2)
				on conflict (username) where (deleted_at is null) do update
				set email = excluded.email,
					updated_at = now()
				returning id
			`),
			q,
		)
		assert.EqualValues(t,
			[]any{
				"tester1", "tester1@localhost",
			},
			args,
		)
	})

	t.Run("insert on conflict index do update", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "email")
			b.Value("tester1", "tester1@localhost")
			b.OnConflictIndex("username").DoUpdate(func(b pgstmt.UpdateStatement) {
				b.Set("email").ToRaw("excluded.email")
				b.Set("updated_at").ToRaw("now()")
			})
			b.Returning("id")
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				insert into users (username, email)
				values ($1, $2)
				on conflict (username) do update
				set email = excluded.email,
					updated_at = now()
				returning id
			`),
			q,
		)
		assert.EqualValues(t,
			[]any{
				"tester1", "tester1@localhost",
			},
			args,
		)
	})

	t.Run("insert on conflict on constraint do update", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "email")
			b.Value("tester1", "tester1@localhost")
			b.OnConflictOnConstraint("username_key").DoUpdate(func(b pgstmt.UpdateStatement) {
				b.Set("email").ToRaw("excluded.email")
				b.Set("updated_at").ToRaw("now()")
			})
			b.Returning("id")
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				insert into users (username, email)
				values ($1, $2)
				on conflict on constraint username_key do update
				set email = excluded.email,
					updated_at = now()
				returning id
			`),
			q,
		)
		assert.EqualValues(t,
			[]any{
				"tester1", "tester1@localhost",
			},
			args,
		)
	})

	t.Run("values", func(t *testing.T) {
		q, args := pgstmt.Insert(func(b pgstmt.InsertStatement) {
			b.Into("users")
			b.Columns("username", "name")
			b.Values([][]any{
				{"tester1", "Tester 1"},
				{"tester2", "Tester 2"},
			}...)
		}).SQL()

		assert.Equal(t,
			stripSpace(`
				insert into users (username, name)
				values ($1, $2),
				       ($3, $4)
            `),
			q,
		)
		assert.EqualValues(t,
			[]any{
				"tester1", "Tester 1",
				"tester2", "Tester 2",
			},
			args,
		)
	})
}
