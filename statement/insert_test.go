package statement_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/acoshift/pgsql/statement"
)

func TestInsertIntoStatement(t *testing.T) {
	t.Parallel()

	s := InsertInto("table1")
	s.Columns("id", "username", "created_at")
	s.Values(1, "test", "now()")
	s.Returning("id")

	assert.Equal(t,
		"insert into table1 (id, username, created_at) values ($1, $2, $3) returning id;",
		s.QueryString(),
	)
	assert.EqualValues(t,
		[]interface{}{1, "test", "now()"},
		s.Args(),
	)
}
