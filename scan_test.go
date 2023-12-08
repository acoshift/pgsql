package pgsql_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/acoshift/pgsql"
)

func TestScan(t *testing.T) {
	t.Parallel()

	db := open(t)
	defer db.Close()

	_, err := db.Exec(`
		drop table if exists test_pgsql_scan;
		create table test_pgsql_scan (
			id int primary key,
			json_value json,
			array_value bigint[]
		);
		insert into test_pgsql_scan (id, json_value, array_value) values (1, '{"a": "test", "b": 7}', '{1, 2, 3}');
	`)
	if !assert.NoError(t, err) {
		return
	}
	defer db.Exec(`drop table test_pgsql_scan`)

	var obj struct {
		A string
		B int
	}
	var arr []int64

	err = pgsql.Scan(db.QueryRow(`
		select json_value, array_value
		from test_pgsql_scan
		where id = 1
	`).Scan)(&obj, &arr)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "test", obj.A)
	assert.Equal(t, 7, obj.B)
	assert.Equal(t, []int64{1, 2, 3}, arr)
}
