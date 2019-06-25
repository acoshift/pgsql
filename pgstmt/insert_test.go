package pgstmt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql/pgstmt"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	q, args := pgstmt.Insert(func(b *pgstmt.InsertBuilder) {
		b.Into("users")
		b.Columns("username", "name", "created_at")
		b.Value("tester1", "Tester 1", "now()")
		b.Value("tester2", "Tester 2", "now()")
		b.Returning("id", "name")
	}).SQL()

	assert.Equal(t,
		"insert into users (username, name, created_at) values ($1, $2, $3), ($4, $5, $6) returning id, name",
		q,
	)
	assert.EqualValues(t,
		[]interface{}{
			"tester1", "Tester 1", "now()",
			"tester2", "Tester 2", "now()",
		},
		args,
	)
}
