package runtime

import (
	"fmt"
	"html"
	"strings"
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

// WriteEscaped writes an escaped value into the builder.
func WriteEscaped(b *strings.Builder, v any) {
	if b == nil || v == nil {
		return
	}
	switch t := v.(type) {
	case SafeString:
		b.WriteString(string(t))
	default:
		b.WriteString(html.EscapeString(Stringify(v)))
	}
}

// WriteRaw writes a raw value into the builder.
func WriteRaw(b *strings.Builder, v any) {
	if b == nil || v == nil {
		return
	}
	b.WriteString(Stringify(v))
}
