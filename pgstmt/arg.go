package pgstmt

// Arg marks value as argument to replace with $? when build query
func Arg(v interface{}) interface{} {
	if _, ok := v.(arg); ok {
		return v
	}
	if _, ok := v.(notArg); ok {
		return v
	}
	if _, ok := v.(any); ok {
		return v
	}
	switch v.(type) {
	default:
		return arg{v}
	case arg:
	case notArg:
	case any:
	case defaultValue:
	}
	return v
}

type arg struct {
	value interface{}
}

// NotArg marks value as non-argument
func NotArg(v interface{}) interface{} {
	if _, ok := v.(notArg); ok {
		return v
	}
	return notArg{v}
}

type notArg struct {
	value interface{}
}

// Any marks value as any($?)
func Any(v interface{}) interface{} {
	return any{v}
}

type any struct {
	value interface{}
}

// Default use for insert default value
var Default interface{} = defaultValue{}

type defaultValue struct{}
