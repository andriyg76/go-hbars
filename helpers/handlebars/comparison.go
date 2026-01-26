package handlebars

import (
	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/runtime"
)

// Eq checks if two values are equal.
func Eq(ctx *runtime.Context, args []any) (any, error) {
	a := helpers.GetArg(args, 0)
	b := helpers.GetArg(args, 1)
	return a == b, nil
}

// Ne checks if two values are not equal.
func Ne(ctx *runtime.Context, args []any) (any, error) {
	a := helpers.GetArg(args, 0)
	b := helpers.GetArg(args, 1)
	return a != b, nil
}

// Lt checks if the first value is less than the second.
func Lt(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return false, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return false, nil
	}
	return a < b, nil
}

// Lte checks if the first value is less than or equal to the second.
func Lte(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return false, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return false, nil
	}
	return a <= b, nil
}

// Gt checks if the first value is greater than the second.
func Gt(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return false, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return false, nil
	}
	return a > b, nil
}

// Gte checks if the first value is greater than or equal to the second.
func Gte(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return false, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return false, nil
	}
	return a >= b, nil
}

// And returns true if all arguments are truthy.
func And(ctx *runtime.Context, args []any) (any, error) {
	for _, arg := range args {
		if !helpers.IsTruthy(arg) {
			return false, nil
		}
	}
	return true, nil
}

// Or returns true if any argument is truthy.
func Or(ctx *runtime.Context, args []any) (any, error) {
	for _, arg := range args {
		if helpers.IsTruthy(arg) {
			return true, nil
		}
	}
	return false, nil
}

// Not returns the negation of a value.
func Not(ctx *runtime.Context, args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	return !helpers.IsTruthy(arg), nil
}

