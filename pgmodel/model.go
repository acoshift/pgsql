package pgmodel

import (
	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgstmt"
)

// Scanner model
type Scanner interface {
	Scan(scan pgsql.Scanner) error
}

// Selector model
type Selector interface {
	Select(b pgstmt.SelectStatement)
	Scanner
}

// Inserter model
type Inserter interface {
	Insert(b pgstmt.InsertStatement)
}

// Updater model
type Updater interface {
	Update(b pgstmt.UpdateStatement)
}
