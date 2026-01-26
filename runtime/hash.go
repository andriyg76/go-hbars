package runtime

import "fmt"

// Hash represents named helper arguments.
type Hash map[string]any

// HashArg returns the trailing Hash argument if present.
func HashArg(args []any) (Hash, bool) {
	if len(args) == 0 {
		return nil, false
	}
	switch v := args[len(args)-1].(type) {
	case Hash:
		return v, true
	case map[string]any:
		return Hash(v), true
	default:
		return nil, false
	}
}

// MissingPartial formats a missing partial error.
func MissingPartial(name string) error {
	return fmt.Errorf("partial %q is not defined", name)
}
