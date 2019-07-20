package pgstmt_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
	"github.com/acoshift/pgsql/pgstmt"
)

func TestResult_QueryAllWith(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = pgctx.NewContext(ctx, db)

	type v struct {
		a int
		b int
	}

	var vs []*v
	err := pgstmt.Select(func(b pgstmt.SelectStatement) {
		b.Columns("*")
		b.FromValues(func(b pgstmt.Values) {
			b.Value("1", "2")
			b.Value("3", "4")
			b.Value("5", "6")
		}, "t")
	}).IterWith(ctx, func(scan pgsql.Scanner) error {
		var x v
		err := scan(&x.a, &x.b)
		if err != nil {
			return err
		}
		vs = append(vs, &x)
		return nil
	})

	assert.NoError(t, err)

	assert.Len(t, vs, 3)
	assert.EqualValues(t, []*v{
		{1, 2},
		{3, 4},
		{5, 6},
	}, vs)
}
