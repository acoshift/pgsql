package pgstmt

// Insert builds insert statement
func Insert(f func(b InsertStatement)) *Result {
	var st insertStmt
	f(&st)

	var b buffer
	b.push("insert")
	if st.table != "" {
		b.push("into", st.table)
	}
	if !st.columns.empty() {
		b.push(&st.columns)
	}
	if st.overridingValue != "" {
		b.push("overriding")
		b.push(st.overridingValue)
		b.push("value")
	}
	if st.defaultValues {
		b.push("default values")
	}
	if !st.values.empty() {
		b.push("values")
		b.push(&st.values)
	}
	if st.conflict != nil {
		b.push("on conflict")
		if len(st.conflict.targets) > 0 {
			var p parenGroup
			p.pushString(st.conflict.targets...)
			b.push(&p)
		}
		if st.conflict.doNothing {
			b.push("do nothing")
		}
	}
	if !st.returning.empty() {
		b.push("returning")
		b.push(&st.returning)
	}

	return newResult(build(&b))
}

// InsertStatement is the insert statement builder
type InsertStatement interface {
	Into(table string)
	Columns(col ...string)
	OverridingSystemValue()
	OverridingUserValue()
	DefaultValues()
	Value(value ...interface{})
	Values(values ...interface{})
	OnConflict(target ...string) ConflictAction
	Returning(col ...string)
}

type ConflictAction interface {
	DoNothing()
	// DoUpdate(f func())
}

type insertStmt struct {
	table           string
	columns         parenGroup
	overridingValue string
	defaultValues   bool
	conflict        *conflictAction
	values          group
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

func (st *insertStmt) Value(value ...interface{}) {
	var x parenGroup
	for _, v := range value {
		x.push(Arg(v))
	}
	st.values.push(&x)
}

func (st *insertStmt) Values(values ...interface{}) {
	for _, value := range values {
		st.Value(value)
	}
}

func (st *insertStmt) OnConflict(target ...string) ConflictAction {
	st.conflict = &conflictAction{targets: target}
	return st.conflict
}

func (st *insertStmt) Returning(col ...string) {
	st.returning.pushString(col...)
}

type conflictAction struct {
	targets   []string
	doNothing bool
	// TODO: add doUpdate
}

func (b *conflictAction) DoNothing() {
	b.doNothing = true
}
