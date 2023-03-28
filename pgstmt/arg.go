package pgstmt

// Arg marks value as argument to replace with $? when build query
func Arg(v any) any {
	switch v.(type) {
	default:
		return arg{v}
	case arg:
	case notArg:
	case raw:
	case _any:
	case all:
	case defaultValue:
	}
	return v
}

type arg struct {
	value any
}

// NotArg marks value as non-argument
func NotArg(v any) any {
	if _, ok := v.(notArg); ok {
		return v
	}
	return notArg{v}
}

type notArg struct {
	value any
}

// Raw marks value as raw sql without escape
func Raw(v any) any {
	switch v := v.(type) {
	default:
		return raw{v}
	case _any:
		return Any(raw{v.value})
	}
}

type raw struct {
	value any
}

// Any marks value as any($?)
func Any(v any) any {
	return _any{v}
}

type _any struct {
	value any
}

// All marks value as all($?)
func All(v any) any {
	return all{v}
}

type all struct {
	value any
}

// Default use for insert default value
var Default any = defaultValue{}

type defaultValue struct{}
