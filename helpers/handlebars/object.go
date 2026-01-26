package handlebars

import (
	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/runtime"
)

// Has returns true if an object has a property.
func Has(ctx *runtime.Context, args []any) (any, error) {
	obj := helpers.GetArg(args, 0)
	key := helpers.GetStringArg(args, 1)
	
	if key == "" {
		return false, nil
	}
	
	switch v := obj.(type) {
	case map[string]any:
		_, ok := v[key]
		return ok, nil
	case map[any]any:
		_, ok := v[key]
		return ok, nil
	}
	return false, nil
}

// Keys returns the keys of an object.
func Keys(ctx *runtime.Context, args []any) (any, error) {
	obj := helpers.GetArg(args, 0)
	
	switch v := obj.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys, nil
	case map[any]any:
		keys := make([]any, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys, nil
	}
	return []any{}, nil
}

// Values returns the values of an object.
func Values(ctx *runtime.Context, args []any) (any, error) {
	obj := helpers.GetArg(args, 0)
	
	switch v := obj.(type) {
	case map[string]any:
		values := make([]any, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values, nil
	case map[any]any:
		values := make([]any, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values, nil
	}
	return []any{}, nil
}

// Size returns the size of an object or array.
func Size(ctx *runtime.Context, args []any) (any, error) {
	return Length(ctx, args)
}

// IsEmpty checks if a value is empty.
func IsEmpty(ctx *runtime.Context, args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	return helpers.IsEmpty(arg), nil
}

// IsNotEmpty checks if a value is not empty.
func IsNotEmpty(ctx *runtime.Context, args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	return !helpers.IsEmpty(arg), nil
}

