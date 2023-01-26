package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestUnion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		result *pgstmt.Result
		query  string
		args   []interface{}
	}{
		{
			"union select",
			pgstmt.Union(func(b pgstmt.UnionStatement) {
				b.Select(func(b pgstmt.SelectStatement) {
					b.Columns("id")
					b.From("table1")
				})
				b.AllSelect(func(b pgstmt.SelectStatement) {
					b.Columns("id")
					b.From("table2")
				})
				b.OrderBy("id")
				b.Limit(10)
				b.Offset(2)
			}),
			`
				(select id from table1)
				union all (select id from table2)
				order by id
				limit 10 offset 2
			`,
			nil,
		},
		{
			"union nested",
			pgstmt.Union(func(b pgstmt.UnionStatement) {
				b.Union(func(b pgstmt.UnionStatement) {
					b.Select(func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table1")
					})
					b.Select(func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table2")
					})
				})
				b.Select(func(b pgstmt.SelectStatement) {
					b.Columns("id")
					b.From("table3")
				})
				b.AllUnion(func(b pgstmt.UnionStatement) {
					b.Select(func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table4")
					})
					b.Select(func(b pgstmt.SelectStatement) {
						b.Columns("id")
						b.From("table5")
					})
				})
			}),
			`
				(
					(select id from table1)
					union (select id from table2)
				)
				union (select id from table3)
				union all (
					(select id from table4)
					union
					(select id from table5)
				)
			`,
			nil,
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
