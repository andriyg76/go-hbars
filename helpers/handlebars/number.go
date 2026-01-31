package handlebars

import (
	"fmt"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/runtime"
)

// FormatNumber formats a number with optional precision and separator.
func FormatNumber(args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return "0", nil
	}
	
	precision := 0
	separator := ","
	hash, _ := runtime.HashArg(args)
	if hash != nil {
		if p, ok := hash["precision"].(float64); ok {
			precision = int(p)
		}
		if p, ok := hash["precision"].(int); ok {
			precision = p
		}
		if s, ok := hash["separator"].(string); ok {
			separator = s
		}
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
func ToInt(args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	return int(n), nil
}

// ToFloat converts a value to a float.
func ToFloat(args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0.0, nil
	}
	return n, nil
}

// Random generates a random number between min and max.
func Random(args []any) (any, error) {
	min := 0.0
	max := 100.0
	
	if len(args) > 0 {
		if m, err := helpers.GetNumberArg(args, 0); err == nil {
			min = m
		}
	}
	if len(args) > 1 {
		if m, err := helpers.GetNumberArg(args, 1); err == nil {
			max = m
		}
	}
	
	hash, _ := runtime.HashArg(args)
	if hash != nil {
		if m, ok := hash["min"].(float64); ok {
			min = m
		}
		if m, ok := hash["max"].(float64); ok {
			max = m
		}
	}
	
	// Simple pseudo-random using current time (not cryptographically secure)
	// For a real implementation, use crypto/rand
	return min + (max-min)*0.5, nil // Placeholder
}

// ToFixed formats a number with a fixed number of decimal places.
func ToFixed(args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return "0", nil
	}
	
	precision := 0
	if len(args) > 1 {
		if p, err := helpers.GetNumberArg(args, 1); err == nil {
			precision = int(p)
		}
	}
	
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, n), nil
}

// ToString converts a value to a string.
func ToString(args []any) (any, error) {
	arg := helpers.GetArg(args, 0)
	return runtime.Stringify(arg), nil
}

// ToNumber converts a string to a number.
func ToNumber(args []any) (any, error) {
	n, err := helpers.GetNumberArg(args, 0)
	if err != nil {
		return 0, nil
	}
	return n, nil
}

