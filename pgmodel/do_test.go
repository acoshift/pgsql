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

func TestDo_UpdateModel(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = pgctx.NewContext(ctx, db)

	_, err := db.Exec(`
		drop table if exists test_pgmodel_update;
		create table test_pgmodel_update (
			id int primary key,
			value varchar not null,
			created_at timestamptz not null default now(),
			updated_at timestamptz
		);
		insert into test_pgmodel_update (id, value)
		values (1, 'value 1'),
			   (2, 'value 2');
	`)
	assert.NoError(t, err)

	{
		err = pgmodel.Do(ctx, &updateModel{Value: "new value"}, pgmodel.Equal("id", 1))
		assert.NoError(t, err)

		var m updateSelectModel
		err = pgmodel.Do(ctx, &m, pgmodel.Equal("id", 1))
		assert.NoError(t, err)
		assert.Equal(t, int64(1), m.ID)
		assert.Equal(t, "new value", m.Value)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	}

	{
		u := updateModelWithReturn{Value: "update value"}
		err = pgmodel.Do(ctx, &u, pgmodel.Equal("id", 2))
		assert.NoError(t, err)

		var m updateSelectModel
		err = pgmodel.Do(ctx, &m, pgmodel.Equal("id", 2))
		assert.NoError(t, err)
		assert.Equal(t, int64(2), m.ID)
		assert.Equal(t, "update value", m.Value)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
		assert.Equal(t, m.UpdatedAt, u.Return.UpdatedAt)
	}
}

type updateModel struct {
	Value string
}

func (m *updateModel) Update(b pgstmt.UpdateStatement) {
	b.Table("test_pgmodel_update")
	b.Set("value").To(m.Value)
	b.Set("updated_at").ToRaw("now()")
}

type updateModelWithReturn struct {
	Value string

	Return struct {
		UpdatedAt time.Time
	}
}

func (m *updateModelWithReturn) Update(b pgstmt.UpdateStatement) {
	b.Table("test_pgmodel_update")
	b.Set("value").To(m.Value)
	b.Set("updated_at").ToRaw("now()")
	b.Returning("updated_at")
}

func (m *updateModelWithReturn) Scan(scan pgsql.Scanner) error {
	return scan(&m.Return.UpdatedAt)
}

type updateSelectModel struct {
	ID        int64
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *updateSelectModel) Select(b pgstmt.SelectStatement) {
	b.Columns("id", "value", "created_at", "updated_at")
	b.From("test_pgmodel_update")
}

func (m *updateSelectModel) Scan(scan pgsql.Scanner) error {
	return scan(&m.ID, &m.Value, &m.CreatedAt, pgsql.NullTime(&m.UpdatedAt))
}

func TestDo_InsertModel(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	ctx := context.Background()
	ctx = pgctx.NewContext(ctx, db)

	_, err := db.Exec(`
		drop table if exists test_pgmodel_insert;
		create table test_pgmodel_insert (
			id int primary key,
			value varchar not null,
			created_at timestamptz not null default now()
		);
	`)
	assert.NoError(t, err)

	err = pgmodel.Do(ctx, &insertModel{ID: 1, Value: "value 1"})
	assert.NoError(t, err)
}

type insertModel struct {
	ID    int64
	Value string
}

func (m *insertModel) Insert(b pgstmt.InsertStatement) {
	b.Into("test_pgmodel_insert")
	b.Columns("id", "value")
	b.Value(m.ID, m.Value)
}
