package runtime

import "io"

// Helper is a user-defined function invoked from a template.
type Helper func(ctx *Context, args []any) (any, error)

// BlockHelper is a user-defined block helper function.
type BlockHelper func(ctx *Context, args []any, options BlockOptions) error

// BlockOptions contains the render functions for a block helper.
type BlockOptions struct {
	Fn      func(ctx *Context, w io.Writer) error
	Inverse func(ctx *Context, w io.Writer) error
}
