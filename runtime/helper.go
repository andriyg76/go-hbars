package runtime

// Helper is a user-defined function invoked from a template.
type Helper func(ctx *Context, args []any) (any, error)
