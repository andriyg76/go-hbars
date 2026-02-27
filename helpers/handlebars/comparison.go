package handlebars

import (
	"github.com/andriyg76/go-hbars/runtime"
)

// Eq checks if two values are equal.
func Eq(args runtime.HelperArgs) (any, error) {
	a := args.GetArg(0)
	b := args.GetArg(1)
	return a == b, nil
}

// Ne checks if two values are not equal.
func Ne(args runtime.HelperArgs) (any, error) {
	a := args.GetArg(0)
	b := args.GetArg(1)
	return a != b, nil
}

// Lt checks if the first value is less than the second.
func Lt(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return false, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return false, nil
	}
	return a < b, nil
}

// Lte checks if the first value is less than or equal to the second.
func Lte(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return false, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return false, nil
	}
	return a <= b, nil
}

// Gt checks if the first value is greater than the second.
func Gt(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return false, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return false, nil
	}
	return a > b, nil
}

// Gte checks if the first value is greater than or equal to the second.
func Gte(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return false, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return false, nil
	}
	return a >= b, nil
}

// And returns true if all arguments are truthy.
func And(args runtime.HelperArgs) (any, error) {
	for _, arg := range args.Args {
		ok, err := runtime.IsTruthy(arg)
		if err != nil {
			return nil, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// Or returns true if any argument is truthy.
func Or(args runtime.HelperArgs) (any, error) {
	for _, arg := range args.Args {
		ok, err := runtime.IsTruthy(arg)
		if err != nil {
			return nil, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// Not returns the negation of a value.
func Not(args runtime.HelperArgs) (any, error) {
	arg := args.GetArg(0)
	ok, err := runtime.IsTruthy(arg)
	if err != nil {
		return nil, err
	}
	return !ok, nil
}

