package runtime

import (
	"fmt"
	"io"
	"log"
)

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

// MissingPartialOutput is used for dynamic partials only: when the partial name
// is not found, it writes the error message (HTML comment) and logs with
// log.Error; the render continues without failing. If w implements
// LazyBlockWriter, the output is registered as a lazy block and written on Flush.
func MissingPartialOutput(w io.Writer, name string) {
	msg := fmt.Sprintf("partial %q is not defined", name)
	log.Printf("[ERROR] %s", msg)
	htmlComment := "<!-- " + msg + " -->"
	if lw, ok := w.(LazyBlockWriter); ok {
		lw.WriteLazyBlock(func(out io.Writer) {
			io.WriteString(out, htmlComment)
		})
		return
	}
	io.WriteString(w, htmlComment)
}
