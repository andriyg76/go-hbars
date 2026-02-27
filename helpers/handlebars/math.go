package handlebars

import (
	"math"

	"github.com/andriyg76/go-hbars/runtime"
)

// Add adds two numbers.
func Add(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return 0, nil
	}
	return a + b, nil
}

// Subtract subtracts the second number from the first.
func Subtract(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return 0, nil
	}
	return a - b, nil
}

// Multiply multiplies two numbers.
func Multiply(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return 0, nil
	}
	return a * b, nil
}

// Divide divides the first number by the second.
func Divide(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return 0, nil
	}
	if b == 0 {
		return 0, nil
	}
	return a / b, nil
}

// Modulo returns the remainder of dividing the first number by the second.
func Modulo(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return 0, nil
	}
	if b == 0 {
		return 0, nil
	}
	return math.Mod(a, b), nil
}

// Floor returns the floor of a number.
func Floor(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	return math.Floor(n), nil
}

// Ceil returns the ceiling of a number.
func Ceil(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	return math.Ceil(n), nil
}

// Round rounds a number to the nearest integer.
func Round(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	return math.Round(n), nil
}

// Abs returns the absolute value of a number.
func Abs(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	return math.Abs(n), nil
}

// Min returns the minimum of two numbers.
func Min(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return a, nil
	}
	return math.Min(a, b), nil
}

// Max returns the maximum of two numbers.
func Max(args runtime.HelperArgs) (any, error) {
	a, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	b, err := args.GetNumber(1)
	if err != nil {
		return a, nil
	}
	return math.Max(a, b), nil
}

