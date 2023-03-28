package pgstmt_test

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		result *pgstmt.Result
		query  string
		args   []any
	}{
		{
			"only select",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("1")
			}),
			"select 1",
			nil,
		},
		{
			"select arg",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns(pgstmt.Arg("x"))
			}),
			"select $1",
			[]any{
				"x",
			},
		},
		{
			"select without arg",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns(1, "x", 1.2)
			}),
			"select 1, x, 1.2",
			nil,
		},
		{
			"select from",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
			}),
			"select id, name from users",
			nil,
		},
		{
			"select from select",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
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
			}),
			`
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
			`,
			[]any{
				1,
			},
		},
		{
			"select from where",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
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
			}),
			"select id, name from users where (id = $1 and name = $2 and is_active = $3) and ((age = $4) or (age = $5))",
			[]any{
				3,
				"test",
				true,
				15,
				18,
			},
		},
		{
			"select from where order",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.Where(func(b pgstmt.Cond) {
					b.Eq("id", 1)
				})
				b.OrderBy("created_at").Asc().NullsLast()
				b.OrderBy("id").Desc()
			}),
			"select id, name from users where (id = $1) order by created_at asc nulls last, id desc",
			[]any{
				1,
			},
		},
		{
			"select limit offset",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.Where(func(b pgstmt.Cond) {
					b.Eq("id", 1)
				})
				b.OrderBy("id")
				b.Limit(5)
				b.Offset(10)
			}),
			"select id, name from users where (id = $1) order by id limit 5 offset 10",
			[]any{
				1,
			},
		},
		{
			"join",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.LeftJoin("roles using id")
			}),
			"select id, name from users left join roles using id",
			nil,
		},
		{
			"join on",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.LeftJoin("roles").On(func(b pgstmt.Cond) {
					b.EqRaw("users.id", "roles.id")
				})
			}),
			"select id, name from users left join roles on (users.id = roles.id)",
			nil,
		},
		{
			"join using",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name")
				b.From("users")
				b.InnerJoin("roles").Using("id", "name")
			}),
			"select id, name from users inner join roles using (id, name)",
			nil,
		},
		{
			"join select",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id", "name", "count(*)")
				b.From("users")
				b.LeftJoinSelect(func(b pgstmt.SelectStatement) {
					b.Columns("user_id", "data")
					b.From("event")
				}, "t").On(func(b pgstmt.Cond) {
					b.EqRaw("t.user_id", "users.id")
				})
				b.GroupBy("id", "name")
			}),
			"select id, name, count(*) from users left join (select user_id, data from event) t on (t.user_id = users.id) group by (id, name)",
			nil,
		},
		{
			"group by having",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("city", "max(temp_lo)")
				b.From("weather")
				b.GroupBy("city")
				b.Having(func(b pgstmt.Cond) {
					b.LtRaw("max(temp_lo)", 40)
				})
			}),
			"select city, max(temp_lo) from weather group by (city) having (max(temp_lo) < 40)",
			nil,
		},
		{
			"select any",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.Eq("x", pgstmt.Any(pq.Array([]int64{1, 2})))
				})
			}),
			"select * from table where (x = any($1))",
			[]any{
				pq.Array([]int64{1, 2}),
			},
		},
		{
			"select all",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.Ne("x", pgstmt.All(pq.Array([]int64{1, 2})))
				})
			}),
			"select * from table where (x != all($1))",
			[]any{
				pq.Array([]int64{1, 2}),
			},
		},
		{
			"select in",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.In("x", 1, 2)
				})
			}),
			"select * from table where (x in ($1, $2))",
			[]any{
				1,
				2,
			},
		},
		{
			"select in select",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.InSelect("id", func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table2")
					})
				})
			}),
			"select * from table where (id in (select id from table2))",
			nil,
		},
		{
			"select not in",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.NotIn("x", 1, 2)
				})
			}),
			"select * from table where (x not in ($1, $2))",
			[]any{
				1,
				2,
			},
		},
		{
			"select and mode",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.Mode().And()
					b.EqRaw("a", 1)
					b.EqRaw("a", 2)
				})
			}),
			"select * from table where (a = 1 and a = 2)",
			nil,
		},
		{
			"select or mode",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.Mode().Or()
					b.EqRaw("a", 1)
					b.EqRaw("a", 2)
				})
			}),
			"select * from table where (a = 1 or a = 2)",
			nil,
		},
		{
			"select nested or mode",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.EqRaw("a", 1)
					b.And(func(b pgstmt.Cond) {
						b.Mode().Or()
						b.EqRaw("a", 2)
						b.EqRaw("a", 3)
					})
				})
			}),
			"select * from table where (a = 1) and (a = 2 or a = 3)",
			nil,
		},
		{
			"select nested and",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.EqRaw("a", 1)
					b.EqRaw("b", 1)
					b.And(func(b pgstmt.Cond) {
						b.And(func(b pgstmt.Cond) {
							b.EqRaw("c", 1)
							b.EqRaw("d", 1)
						})
						b.Or(func(b pgstmt.Cond) {
							b.EqRaw("e", 1)
							b.EqRaw("f", 1)
						})
					})
				})
			}),
			"select * from table where (a = 1 and b = 1) and ((c = 1 and d = 1) or (e = 1 and f = 1))",
			nil,
		},
		{
			"select nested and single or without ops",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.EqRaw("a", 1)
					b.EqRaw("b", 1)
					b.And(func(b pgstmt.Cond) {
						// nothing to `or` with
						b.Or(func(b pgstmt.Cond) {
							b.EqRaw("c", 1)
							b.EqRaw("d", 1)
						})
					})
				})
			}),
			"select * from table where (a = 1 and b = 1) and (c = 1 and d = 1)",
			nil,
		},
		{
			"select without op but nested",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table")
				b.Where(func(b pgstmt.Cond) {
					b.And(func(b pgstmt.Cond) {
						b.Mode().Or()
						b.EqRaw("a", 2)
						b.EqRaw("a", 3)
					})
				})
			}),
			"select * from table where (a = 2 or a = 3)",
			nil,
		},
		{
			"select distinct",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Distinct()
				b.Columns("col_1")
			}),
			"select distinct col_1",
			nil,
		},
		{
			"select distinct on",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Distinct().On("col_1", "col_2")
				b.Columns("col_1", "col_3")
			}),
			"select distinct on (col_1, col_2) col_1, col_3",
			nil,
		},
		{
			"left join lateral",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("m.name")
				b.From("manufacturers m")
				b.LeftJoin("lateral get_product_names(m.id) pname").On(func(b pgstmt.Cond) {
					b.Raw("true")
				})
				b.Where(func(b pgstmt.Cond) {
					b.IsNull("pname")
				})
			}),
			`
				select m.name
				from manufacturers m left join lateral get_product_names(m.id) pname on (true)
				where (pname is null)
			`,
			nil,
		},
		{
			"left join lateral select",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("m.name")
				b.From("manufacturers m")
				b.LeftJoinLateralSelect(func(b pgstmt.SelectStatement) {
					b.Columns("get_product_names(m.id) pname")
				}, "t").On(func(b pgstmt.Cond) {
					b.Raw("true")
				})
				b.Where(func(b pgstmt.Cond) {
					b.IsNull("pname")
				})
			}),
			`
				select m.name
				from manufacturers m left join lateral (select get_product_names(m.id) pname) t on (true)
				where (pname is null)
			`,
			nil,
		},
		{
			"inner join union",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("id")
				b.From("table1")
				b.InnerJoinUnion(func(b pgstmt.UnionStatement) {
					b.Select(func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table2")
					})
					b.AllSelect(func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table3")
					})
					b.OrderBy("id").Desc()
					b.Limit(100)
				}, "t").Using("id")
			}),
			`
				select id
				from table1
				inner join (
					(select id from table2)
					union all
					(select id from table3)
					order by id desc
					limit 100
				) t using (id)
			`,
			nil,
		},
		{
			"select where not",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table1")
				b.Where(func(b pgstmt.Cond) {
					b.Eq("id", 1)
					b.Not(func(b pgstmt.Cond) {
						b.Op("tags", "@>", pq.Array([]string{"a", "b"}))
					})
				})
			}),
			`
				select *
				from table1
				where (id = $1 and (not (tags @> $2)))
			`,
			[]any{1, pq.Array([]string{"a", "b"})},
		},
		{
			"select cond eq",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table1")
				b.Where(func(b pgstmt.Cond) {
					b.Field("id").Eq().Value(1)
					b.Field("name").Eq().Field("old_name")
					b.Value(2).Eq().Field(pgstmt.Any("path"))
					b.Field("t1").In().Value(3, 4)
					b.Field("t2").In().Select(func(b pgstmt.SelectStatement) {
						b.Columns(1)
					})
					b.Field("deleted_at").IsNull()
				})
			}),
			`
				select *
				from table1
				where (id = $1
				   and name = old_name
				   and $2 = any(path)
				   and t1 in ($3, $4)
				   and t2 in (select 1)
				   and deleted_at is null)
			`,
			[]any{1, 2, 3, 4},
		},
		{
			"select where any",
			pgstmt.Select(func(b pgstmt.SelectStatement) {
				b.Columns("*")
				b.From("table1")
				b.Where(func(b pgstmt.Cond) {
					b.Eq(pgstmt.Arg(1), pgstmt.Any(pgstmt.Raw("path")))
				})
			}),
			`
				select *
				from table1
				where ($1 = any(path))
			`,
			[]any{1},
		},
	}

	for _, tC := range cases {
		t.Run(tC.name, func(t *testing.T) {
			q, args := tC.result.SQL()
			assert.Equal(t, stripSpace(tC.query), q)
			assert.EqualValues(t, tC.args, args)
		})
	}
}
