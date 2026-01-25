package runtime

import (
	"strconv"
	"strings"
)

// ArgKind describes a helper argument kind.
type ArgKind int

const (
	ArgPath ArgKind = iota
	ArgString
	ArgNumber
	ArgBool
	ArgNull
)

// EvalArg resolves a helper argument in a context.
func EvalArg(ctx *Context, kind ArgKind, value string) any {
	switch kind {
	case ArgPath:
		v, _ := ResolvePath(ctx, value)
		return v
	case ArgString:
		return value
	case ArgNumber:
		return parseNumber(value)
	case ArgBool:
		return value == "true"
	case ArgNull:
		return nil
	default:
		return nil
	}
}

func parseNumber(value string) any {
	if value == "" {
		return float64(0)
	}
	if !strings.ContainsAny(value, ".eE") {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	return float64(0)
}
