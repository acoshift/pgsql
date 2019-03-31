package statement_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/acoshift/pgsql/statement"
)

func TestDeleteFromStatement(t *testing.T) {
	t.Parallel()

	t.Run("Delete all records", func(t *testing.T) {
		s := DeleteFrom("table1")
		assert.Equal(t,
			"delete from table1",
			s.QueryString(),
		)
		assert.Empty(t, s.Args())
	})
}
