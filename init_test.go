package pgsql_test

import (
	"database/sql"
	"testing"
)

func open(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", "postgres://postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatalf("open database connection error; %v", err)
	}
	return db
}
