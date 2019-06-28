package pgstmt

// Arg marks value as argument to replace with $? when build query
func Arg(v interface{}) interface{} {
	if _, ok := v.(arg); ok {
		return v
	}
	if _, ok := v.(notArg); ok {
		return v
	}
	return arg{v}
}

// NotArg marks value as non-argument
func NotArg(v interface{}) interface{} {
	if _, ok := v.(notArg); ok {
		return v
	}
	return notArg{v}
}

type arg struct {
	value interface{}
}

type notArg struct {
	value interface{}
}
