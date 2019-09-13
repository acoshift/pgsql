package pgmodel_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgctx"
	"github.com/acoshift/pgsql/pgmodel"
	"github.com/acoshift/pgsql/pgstmt"
)

func TestDo_SelectModel(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = pgctx.NewContext(ctx, db)

	_, err := db.Exec(`
		drop table if exists test_pgmodel_select;
		create table test_pgmodel_select (
			id int primary key,
			value varchar not null,
			created_at timestamptz not null default now()
		);
		insert into test_pgmodel_select (id, value)
		values (1, 'value 1'),
			   (2, 'value 2');
	`)
	assert.NoError(t, err)

	{
		var m selectModel
		err = pgmodel.Do(ctx, &m, pgmodel.Equal("id", 2))
		assert.NoError(t, err)
		assert.Equal(t, int64(2), m.ID)
		assert.Equal(t, "value 2", m.Value)
		assert.NotEmpty(t, m.CreatedAt)
	}

	{
		var m selectModel
		err = pgmodel.Do(ctx, &m, pgmodel.Equal("id", 99))
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Empty(t, m)
	}

	{
		var ms []*selectModel
		err = pgmodel.Do(ctx, &ms, pgmodel.OrderBy("id desc"), pgmodel.Limit(2))
		assert.NoError(t, err)
		if assert.Len(t, ms, 2) {
			assert.Equal(t, int64(2), ms[0].ID)
			assert.Equal(t, int64(1), ms[1].ID)
		}
	}
}

type selectModel struct {
	ID        int64
	Value     string
	CreatedAt time.Time
}

func (m *selectModel) Select(b pgstmt.SelectStatement) {
	b.Columns("id", "value", "created_at")
	b.From("test_pgmodel_select")
}

func (m *selectModel) Scan(scan pgsql.Scanner) error {
	return scan(&m.ID, &m.Value, &m.CreatedAt)
}
