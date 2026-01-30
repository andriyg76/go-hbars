package runtime

import "io"

// Helper is a user-defined function invoked from a template.
// It receives only args (no context); values are resolved by the compiler.
type Helper func(args []any) (any, error)

// BlockHelper is a user-defined block helper function.
type BlockHelper func(args []any, options BlockOptions) error

// BlockOptions contains the render functions for a block helper.
// Fn and Inverse receive only w; the block body closes over typed context.
type BlockOptions struct {
	Fn      func(w io.Writer) error
	Inverse func(w io.Writer) error
}

// GetBlockOptions extracts BlockOptions from the last element of args.
// Returns (nil, false) if args is empty or the last element is not BlockOptions.
func GetBlockOptions(args []any) (BlockOptions, bool) {
	if len(args) == 0 {
		return BlockOptions{}, false
	}
	if opts, ok := args[len(args)-1].(BlockOptions); ok {
		return opts, true
	}
	return BlockOptions{}, false
}
