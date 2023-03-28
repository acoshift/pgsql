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

	OnConflict(f func(b ConflictTarget)) ConflictAction

	// OnConflictDoNothing is the shortcut for
	// OnConflict(func(b ConflictTarget) {}).DoNothing()
	OnConflictDoNothing()

	// OnConflictIndex is the shortcut for
	// OnConflict(func(b ConflictTarget) {
	//		b.Index(target...)
	// })
	OnConflictIndex(target ...string) ConflictAction

	// OnConflictOnConstraint is the shortcut for
	// OnConflict(func(b ConflictTarget) {
	//		b.OnConstraint(constraintName)
	// })
	OnConflictOnConstraint(constraintName string) ConflictAction

	Returning(col ...string)
}

type ConflictTarget interface {
	Index(target ...string)
	Where(f func(b Cond))
	OnConstraint(constraintName string)
}

type ConflictAction interface {
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

func (st *insertStmt) OnConflict(f func(b ConflictTarget)) ConflictAction {
	var x conflict
	f(&x)
	st.conflict = &x
	return &st.conflict.action
}

func (st *insertStmt) OnConflictDoNothing() {
	st.conflict = &conflict{
		action: conflictAction{
			doNothing: true,
		},
	}
}

func (st *insertStmt) OnConflictIndex(target ...string) ConflictAction {
	return st.OnConflict(func(b ConflictTarget) {
		b.Index(target...)
	})
}

func (st *insertStmt) OnConflictOnConstraint(constraintName string) ConflictAction {
	return st.OnConflict(func(b ConflictTarget) {
		b.OnConstraint(constraintName)
	})
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
		b.push(st.conflict.make())
	}
	if !st.returning.empty() {
		b.push("returning", &st.returning)
	}

	return &b
}

type conflict struct {
	targets    []string
	where      cond
	constraint string
	action     conflictAction
}

func (st *conflict) make() *buffer {
	// on conflict can be one of
	// => ( { index_column_name | ( index_expression ) } [ COLLATE collation ] [ opclass ] [, ...] ) [ WHERE index_predicate ]
	// => ON CONSTRAINT constraint_name

	var b buffer

	b.push("on conflict")

	if len(st.targets) > 0 {
		b.push(parenString(st.targets...))

		if !st.where.empty() {
			b.push("where", &st.where)
		}
	} else if st.constraint != "" {
		b.push("on constraint", st.constraint)
	}

	b.push(st.action.make())

	return &b
}

func (st *conflict) Index(target ...string) {
	st.targets = append(st.targets, target...)
}

func (st *conflict) Where(f func(b Cond)) {
	f(&st.where)
}

func (st *conflict) OnConstraint(constraintName string) {
	st.constraint = constraintName
}

type conflictAction struct {
	doNothing bool
	doUpdate  *updateStmt
}

func (st *conflictAction) make() *buffer {
	var b buffer
	if st.doNothing {
		b.push("do nothing")
	}
	if st.doUpdate != nil {
		b.push("do", st.doUpdate.make())
	}
	return &b
}

func (st *conflictAction) DoNothing() {
	st.doNothing = true
}

func (st *conflictAction) DoUpdate(f func(b UpdateStatement)) {
	var x updateStmt
	f(&x)
	st.doUpdate = &x
}
