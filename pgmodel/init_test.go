package pgmodel_test

import (
	"database/sql"
	"os"
	"testing"
)

func open(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/postgres?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("open database connection error; %v", err)
	}
	return db
}
