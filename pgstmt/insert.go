package pgstmt

// Insert builds insert statement
func Insert(f func(b InsertStatement)) *Result {
	var st insertStmt
	f(&st)
	return newResult(build(st.make()))
}

// InsertStatement is the insert statement builder
type InsertStatement interface {
	Into(table string)
	Columns(col ...string)
	OverridingSystemValue()
	OverridingUserValue()
	DefaultValues()
	Value(value ...any)
	Values(values ...any)
	Select(f func(b SelectStatement))
	OnConflict(target ...string) OnConflict
	OnConflictOnConstraint(constraintName string) OnConflict
	Returning(col ...string)
}

type OnConflict interface {
	DoNothing()
	DoUpdate(f func(b UpdateStatement))
}

type insertStmt struct {
	table           string
	columns         parenGroup
	overridingValue string
	defaultValues   bool
	conflict        *conflict
	values          group
	selects         *selectStmt
	returning       group
}

func (st *insertStmt) Into(table string) {
	st.table = table
}

func (st *insertStmt) Columns(col ...string) {
	st.columns.pushString(col...)
}

func (st *insertStmt) OverridingSystemValue() {
	st.overridingValue = "system"
}

func (st *insertStmt) OverridingUserValue() {
	st.overridingValue = "user"
}

func (st *insertStmt) DefaultValues() {
	st.defaultValues = true
}

func (st *insertStmt) Value(value ...any) {
	var x parenGroup
	for _, v := range value {
		x.push(Arg(v))
	}
	st.values.push(&x)
}

func (st *insertStmt) Values(values ...any) {
	for _, value := range values {
		st.Value(value)
	}
}

func (st *insertStmt) Select(f func(b SelectStatement)) {
	var x selectStmt
	f(&x)
	st.selects = &x
}

func (st *insertStmt) OnConflict(target ...string) OnConflict {
	st.conflict = &conflict{targets: target}
	return st.conflict
}

func (st *insertStmt) OnConflictOnConstraint(constraintName string) OnConflict {
	st.conflict = &conflict{constraint: constraintName}
	return st.conflict
}

func (st *insertStmt) Returning(col ...string) {
	st.returning.pushString(col...)
}

func (st *insertStmt) make() *buffer {
	var b buffer
	b.push("insert")
	if st.table != "" {
		b.push("into", st.table)
	}
	if !st.columns.empty() {
		b.push(&st.columns)
	}
	if st.overridingValue != "" {
		b.push("overriding", st.overridingValue, "value")
	}
	if st.defaultValues {
		b.push("default values")
	}
	if !st.values.empty() {
		b.push("values", &st.values)
	}
	if st.selects != nil {
		b.push(st.selects.make())
	}
	if st.conflict != nil {
		b.push("on conflict")

		// on conflict can be one of
		// => ( { index_column_name | ( index_expression ) } [ COLLATE collation ] [ opclass ] [, ...] ) [ WHERE index_predicate ]
		// => ON CONSTRAINT constraint_name
		if len(st.conflict.targets) > 0 {
			b.push(parenString(st.conflict.targets...))
		} else if st.conflict.constraint != "" {
			b.push("on constraint")
			b.push(st.conflict.constraint)
		}
		if st.conflict.doNothing {
			b.push("do nothing")
		}
		if st.conflict.doUpdate != nil {
			b.push("do", st.conflict.doUpdate.make())
		}
	}
	if !st.returning.empty() {
		b.push("returning", &st.returning)
	}

	return &b
}

type conflict struct {
	targets    []string
	constraint string
	doNothing  bool
	doUpdate   *updateStmt
}

func (st *conflict) DoNothing() {
	st.doNothing = true
}

func (st *conflict) DoUpdate(f func(b UpdateStatement)) {
	var x updateStmt
	f(&x)
	st.doUpdate = &x
}
