package pgsql_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql"
)

func TestJSON(t *testing.T) {
	db := open(t)

	_, err := db.Exec(`
		drop table if exists test_pgsql_json;
		create table test_pgsql_json (
			id int primary key,
			value json
		);
	`)
	if !assert.NoError(t, err) {
		return
	}
	defer db.Exec(`drop table test_pgsql_json`)

	var obj struct {
		A string
		B int
	}

	obj.A = "test"
	obj.B = 7

	var ok bool
	err = db.QueryRow(`
		insert into test_pgsql_json (id, value)
		values (1, $1)
		returning value is not null
	`, pgsql.JSON(&obj)).Scan(&ok)
	assert.NoError(t, err)
	assert.True(t, ok)

	obj.A = ""
	obj.B = 0
	err = db.QueryRow(`
		select value
		from test_pgsql_json
		where id = 1
	`).Scan(pgsql.JSON(&obj))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "test", obj.A)
	assert.Equal(t, 7, obj.B)

	obj.A = ""
	obj.B = 0
	err = db.QueryRow(`select null`).Scan(pgsql.JSON(&obj))
	assert.NoError(t, err)
}
