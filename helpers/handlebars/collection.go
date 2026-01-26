package handlebars

import (
	"reflect"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/runtime"
)

// Lookup looks up a value from the context or data by key.
func Lookup(ctx *runtime.Context, args []any) (any, error) {
	obj := helpers.GetArg(args, 0)
	key := helpers.GetStringArg(args, 1)
	
	if key == "" {
		return nil, nil
	}
	
	switch v := obj.(type) {
	case map[string]any:
		if val, ok := v[key]; ok {
			return val, nil
		}
	case map[any]any:
		if val, ok := v[key]; ok {
			return val, nil
		}
	}
	
	// Try to resolve as a path in the context
	if ctx != nil {
		if val, ok := runtime.ResolvePath(ctx, key); ok {
			return val, nil
		}
	}
	
	return nil, nil
}

// Default returns the first argument if it's truthy, otherwise returns the default value.
func Default(ctx *runtime.Context, args []any) (any, error) {
	value := helpers.GetArg(args, 0)
	defaultVal := helpers.GetArg(args, 1)
	
	hash, _ := runtime.HashArg(args)
	if hash != nil {
		if def, ok := hash["value"]; ok {
			defaultVal = def
		}
	}
	
	if helpers.IsTruthy(value) {
		return value, nil
	}
	if defaultVal != nil {
		return defaultVal, nil
	}
	return "", nil
}

// Length returns the length of a string, array, or object.
func Length(ctx *runtime.Context, args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	if arg == nil {
		return 0, nil
	}
	
	switch v := arg.(type) {
	case string:
		return len(v), nil
	case []any:
		return len(v), nil
	case []string:
		return len(v), nil
	case map[string]any:
		return len(v), nil
	case map[any]any:
		return len(v), nil
	default:
		// Try reflection for other types
		rv := reflect.ValueOf(arg)
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array, reflect.String:
			return rv.Len(), nil
		}
		return 0, nil
	}
}

// First returns the first element of an array.
func First(ctx *runtime.Context, args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	switch v := arg.(type) {
	case []any:
		if len(v) > 0 {
			return v[0], nil
		}
	case []string:
		if len(v) > 0 {
			return v[0], nil
		}
	}
	return nil, nil
}

// Last returns the last element of an array.
func Last(ctx *runtime.Context, args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	switch v := arg.(type) {
	case []any:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	case []string:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	}
	return nil, nil
}

// InArray checks if a value is in an array.
func InArray(ctx *runtime.Context, args []any) (any, error) {
	value := helpers.GetArg(args, 0)
	arr := helpers.GetArg(args, 1)
	
	switch v := arr.(type) {
	case []any:
		for _, item := range v {
			if item == value {
				return true, nil
			}
		}
	case []string:
		valStr := runtime.Stringify(value)
		for _, item := range v {
			if item == valStr {
				return true, nil
			}
		}
	}
	return false, nil
}

