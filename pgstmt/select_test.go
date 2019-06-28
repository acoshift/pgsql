package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	t.Run("only select", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("1")
		}).SQL()

		assert.Equal(t,
			"select 1",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("select from", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
		}).SQL()

		assert.Equal(t,
			"select id, name from users",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("select from select", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("*")
			b.FromSelect(func(b pgstmt.SelectStatement) {
				b.Columns("p.id", "p.name")
				b.ColumnSelect(func(b pgstmt.SelectStatement) {
					b.Columns(stripSpace(`
						json_build_object('content', coalesce(m.content, ''),
										  'type', coalesce(m.type, 0),
										  'timestamp', m.created_at)
					`))
					b.From("messages m")
					b.Where(func(b pgstmt.Cond) {
						b.EqRaw("m.id", "p.id")
					})
					b.OrderBy("created_at").Desc().NullsFirst()
					b.Limit(1)
					b.Offset(2)
				}, "msg")
				b.From("profile p")
				b.LeftJoin("noti n").On(func(b pgstmt.Cond) {
					b.EqRaw("n.id", "p.id")
					b.Eq("n.user_id", 1)
				})
			}, "t")
		}).SQL()
		assert.Equal(t,
			stripSpace(`
				select *
				from (select p.id, p.name, (select json_build_object('content', coalesce(m.content, ''),
																	 'type', coalesce(m.type, 0),
																	 'timestamp', m.created_at)
											from messages m
											where (m.id = p.id)
											order by created_at desc nulls first
											limit 1
											offset 2) msg
					  from profile p
					  left join noti n on (n.id = p.id and n.user_id = $1)) t
			`),
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				1,
			},
			args,
		)
	})

	t.Run("select from where", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
			b.Where(func(b pgstmt.Cond) {
				b.Eq("id", 3)
				b.Eq("name", "test")
				b.And(func(b pgstmt.Cond) {
					b.Eq("age", 15)
					b.Or(func(b pgstmt.Cond) {
						b.Eq("age", 18)
					})
				})
				b.Eq("is_active", true)
			})
		}).SQL()

		assert.Equal(t,
			"select id, name from users where (id = $1 and name = $2 and is_active = $3) and ((age = $4) or (age = $5))",
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				3,
				"test",
				true,
				15,
				18,
			},
			args,
		)
	})

	t.Run("select from where order", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
			b.Where(func(b pgstmt.Cond) {
				b.Eq("id", 1)
			})
			b.OrderBy("created_at").Asc().NullsLast()
			b.OrderBy("id").Desc()
		}).SQL()

		assert.Equal(t,
			"select id, name from users where (id = $1) order by created_at asc nulls last, id desc",
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				1,
			},
			args,
		)
	})

	t.Run("select limit offset", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
			b.Where(func(b pgstmt.Cond) {
				b.Eq("id", 1)
			})
			b.OrderBy("id")
			b.Limit(5)
			b.Offset(10)
		}).SQL()

		assert.Equal(t,
			"select id, name from users where (id = $1) order by id limit 5 offset 10",
			q,
		)
		assert.EqualValues(t,
			[]interface{}{
				1,
			},
			args,
		)
	})

	t.Run("join", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
			b.LeftJoin("roles using id")
		}).SQL()

		assert.Equal(t,
			"select id, name from users left join roles using id",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("join on", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
			b.LeftJoin("roles").On(func(b pgstmt.Cond) {
				b.EqRaw("users.id", "roles.id")
			})
		}).SQL()

		assert.Equal(t,
			"select id, name from users left join roles on (users.id = roles.id)",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("join using", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("id", "name")
			b.From("users")
			b.InnerJoin("roles").Using("id", "name")
		}).SQL()

		assert.Equal(t,
			"select id, name from users inner join roles using (id, name)",
			q,
		)
		assert.Empty(t, args)
	})

	t.Run("group by having", func(t *testing.T) {
		q, args := pgstmt.Select(func(b pgstmt.SelectStatement) {
			b.Columns("city", "max(temp_lo)")
			b.From("weather")
			b.GroupBy("city")
			b.Having(func(b pgstmt.Cond) {
				b.LtRaw("max(temp_lo)", 40)
			})
		}).SQL()

		assert.Equal(t,
			"select city, max(temp_lo) from weather group by (city) having (max(temp_lo) < 40)",
			q,
		)
		assert.Empty(t, args)
	})
}
