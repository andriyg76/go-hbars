package handlebars

import (
	"math"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/runtime"
)

// Add adds two numbers.
func Add(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return 0, nil
	}
	return a + b, nil
}

// Subtract subtracts the second number from the first.
func Subtract(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return 0, nil
	}
	return a - b, nil
}

// Multiply multiplies two numbers.
func Multiply(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return 0, nil
	}
	return a * b, nil
}

// Divide divides the first number by the second.
func Divide(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return 0, nil
	}
	if b == 0 {
		return 0, nil
	}
	return a / b, nil
}

// Modulo returns the remainder of dividing the first number by the second.
func Modulo(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return 0, nil
	}
	if b == 0 {
		return 0, nil
	}
	return math.Mod(a, b), nil
}

// Floor returns the floor of a number.
func Floor(ctx *runtime.Context, args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	return math.Floor(n), nil
}

// Ceil returns the ceiling of a number.
func Ceil(ctx *runtime.Context, args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	return math.Ceil(n), nil
}

// Round rounds a number to the nearest integer.
func Round(ctx *runtime.Context, args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	return math.Round(n), nil
}

// Abs returns the absolute value of a number.
func Abs(ctx *runtime.Context, args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	return math.Abs(n), nil
}

// Min returns the minimum of two numbers.
func Min(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return a, nil
	}
	return math.Min(a, b), nil
}

// Max returns the maximum of two numbers.
func Max(ctx *runtime.Context, args []any) (any, error) {
	a, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	b, err := helpers.GetNumberArg(args, 1)
	if err != nil {
		return a, nil
	}
	return math.Max(a, b), nil
}

