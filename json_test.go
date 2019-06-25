package pgsql_test

import (
	"testing"

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
	if err != nil {
		t.Fatalf("prepare table error; %v", err)
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
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
	if !ok {
		t.Fatalf("invalid object valuer")
	}

	obj.A = ""
	obj.B = 0
	err = db.QueryRow(`
		select value
		from test_pgsql_json
		where id = 1
	`).Scan(pgsql.JSON(&obj))
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
	if obj.A != "test" || obj.B != 7 {
		t.Fatal("invalid object scanner")
	}

	obj.A = ""
	obj.B = 0
	err = db.QueryRow(`select null`).Scan(pgsql.JSON(&obj))
	if err != nil {
		t.Fatalf("sql error; %v", err)
	}
}
