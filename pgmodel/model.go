package pgmodel

import (
	"github.com/acoshift/pgsql"
	"github.com/acoshift/pgsql/pgstmt"
)

// Selector model
type Selector interface {
	Select(b pgstmt.SelectStatement)
	Scan(scan pgsql.Scanner) error
}

// // Inserter model
// type Inserter interface {
// 	Table() string
// 	Columns() []string
// 	Value(value ...interface{})
// }
