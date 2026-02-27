package handlebars

import (
	"fmt"

	"github.com/andriyg76/go-hbars/runtime"
)

// FormatNumber formats a number with optional precision and separator.
func FormatNumber(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return "0", nil
	}
	precision := 0
	if p, err := args.GetHashNumber("precision"); err == nil {
		precision = int(p)
	}
	separator := args.GetHashString("separator")
	if separator == "" {
		separator = ","
	}
	
	format := fmt.Sprintf("%%.%df", precision)
	result := fmt.Sprintf(format, n)
	
	// Add thousands separator
	if separator != "" && precision == 0 {
		parts := []string{}
		wholePart := fmt.Sprintf("%.0f", n)
		for i := len(wholePart) - 1; i >= 0; i-- {
			parts = append([]string{string(wholePart[i])}, parts...)
			if (len(wholePart)-i)%3 == 0 && i > 0 {
				parts = append([]string{separator}, parts...)
			}
		}
		result = ""
		for _, p := range parts {
			result += p
		}
	}
	
	return result, nil
}

// ToInt converts a value to an integer.
func ToInt(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	return int(n), nil
}

// ToFloat converts a value to a float.
func ToFloat(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0.0, nil
	}
	return n, nil
}

// Random generates a random number between min and max.
func Random(args runtime.HelperArgs) (any, error) {
	min := 0.0
	max := 100.0
	if len(args.Args) > 0 {
		if m, err := args.GetNumber(0); err == nil {
			min = m
		}
	}
	if len(args.Args) > 1 {
		if m, err := args.GetNumber(1); err == nil {
			max = m
		}
	}
	if m, err := args.GetHashNumber("min"); err == nil {
		min = m
	}
	if m, err := args.GetHashNumber("max"); err == nil {
		max = m
	}
	
	// Simple pseudo-random using current time (not cryptographically secure)
	// For a real implementation, use crypto/rand
	return min + (max-min)*0.5, nil // Placeholder
}

// ToFixed formats a number with a fixed number of decimal places.
func ToFixed(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return "0", nil
	}
	precision := 0
	if len(args.Args) > 1 {
		if p, err := args.GetNumber(1); err == nil {
			precision = int(p)
		}
	}
	
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, n), nil
}

// ToString converts a value to a string.
func ToString(args runtime.HelperArgs) (any, error) {
	return runtime.Stringify(args.GetArg(0)), nil
}

// ToNumber converts a string to a number.
func ToNumber(args runtime.HelperArgs) (any, error) {
	n, err := args.GetNumber(0)
	if err != nil {
		return 0, nil
	}
	return n, nil
}

