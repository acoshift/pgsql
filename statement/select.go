package statement

// Select creates select statement
func Select() *SelectStatement {
	return &SelectStatement{}
}

// SelectStatement type
type SelectStatement struct {
	FromClause
	WhereClause
}
