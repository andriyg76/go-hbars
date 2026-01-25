package runtime

import (
	"fmt"
	"html"
	"io"
)

// SafeString marks a value as pre-escaped HTML.
type SafeString string

// Stringify converts a value to its string representation.
func Stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	case SafeString:
		return string(t)
	case fmt.Stringer:
		return t.String()
	case error:
		return t.Error()
	default:
		return fmt.Sprint(v)
	}
}

// WriteEscaped writes an escaped value into the writer.
func WriteEscaped(w io.Writer, v any) error {
	if w == nil || v == nil {
		return nil
	}
	switch t := v.(type) {
	case SafeString:
		_, err := io.WriteString(w, string(t))
		return err
	default:
		_, err := io.WriteString(w, html.EscapeString(Stringify(v)))
		return err
	}
}

// WriteRaw writes a raw value into the writer.
func WriteRaw(w io.Writer, v any) error {
	if w == nil || v == nil {
		return nil
	}
	_, err := io.WriteString(w, Stringify(v))
	return err
}
