package statement

// FromClause type
type FromClause struct {
	table string
}

func (c *FromClause) From(table string) {
	c.table = table
}

// WhereClause type
type WhereClause struct {
	Cond
	args []interface{}
}

func (c *WhereClause) QueryString() string {
	return ""
}

func (c *WhereClause) Args() []interface{} {
	return c.args
}
