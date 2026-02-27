package runtime

// HelperArgs is the argument bundle passed to every helper.
// Values are resolved by the compiler; the helper receives typed args and hash.
type HelperArgs struct {
	HashArgs  map[string]any   // named (hash) arguments
	Args     []any            // positional arguments
	BlockFn   func() error    // when IsBlock: closure that renders the main block (writer captured by compiler); nil otherwise
	InverseFn func() error    // when IsBlock and else exists: closure that renders the else block; may be nil
	IsBlock   bool            // true for block helper invocations
}

// Helper is a user-defined function invoked from a template.
// It receives HelperArgs; values are resolved by the compiler.
// For block helpers, call args.BlockFn() or args.InverseFn() with no arguments; the writer is already captured in the closure.
type Helper func(args HelperArgs) (any, error)
