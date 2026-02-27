package handlebars

import (
	"log"
	"reflect"

	"github.com/andriyg76/go-hbars/runtime"
	"github.com/andriyg76/hexerr"
)

// rawGetter is implemented by typed context interfaces (e.g. MainContext).
type rawGetter interface{ Raw() any }

// Lookup looks up a value from the context or data by key.
func Lookup(args runtime.HelperArgs) (any, error) {
	obj := args.GetArg(0)
	key := args.GetString(1)
	
	if key == "" {
		return nil, nil
	}
	// Typed context interfaces expose Raw(); use it for map lookup
	if r, ok := obj.(rawGetter); ok {
		obj = r.Raw()
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
	
	return nil, nil
}

// Default returns the first argument if it's truthy, otherwise returns the default value.
func Default(args runtime.HelperArgs) (any, error) {
	value := args.GetArg(0)
	defaultVal := args.GetArg(1)
	if def := args.GetHash("value"); def != nil {
		defaultVal = def
	}

	ok, err := runtime.IsTruthy(value)
	if err != nil {
		return nil, err
	}
	if ok {
		return value, nil
	}
	if defaultVal != nil {
		return defaultVal, nil
	}
	return "", nil
}

// Length returns the length of a string, array, or object.
func Length(args runtime.HelperArgs) (any, error) {
	arg := args.GetArg(0)
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
		// Reflect path: check ReflectUsageLevel
		level := runtime.GetReflectUsageLevel()
		if level == runtime.ReflectError {
			return nil, hexerr.Newf("reflect usage disabled by ReflectUsageLevel=ERROR (Length for type %T)", arg)
		}
		if level == runtime.ReflectWarn {
			log.Printf("[go-hbars] reflect used in Length for type %T", arg)
		}
		rv := reflect.ValueOf(arg)
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array, reflect.String:
			return rv.Len(), nil
		}
		return 0, nil
	}
}

// First returns the first element of an array.
func First(args runtime.HelperArgs) (any, error) {
	arg := args.GetArg(0)
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
func Last(args runtime.HelperArgs) (any, error) {
	arg := args.GetArg(0)
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
func InArray(args runtime.HelperArgs) (any, error) {
	value := args.GetArg(0)
	arr := args.GetArg(1)
	
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

