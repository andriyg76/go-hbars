package runtime

import (
	"errors"
	"strings"
	"testing"
)

type testStringer struct{}

func (testStringer) String() string { return "stringer" }

func TestStringify(t *testing.T) {
	if got := Stringify(nil); got != "" {
		t.Fatalf("Stringify nil = %q", got)
	}
	if got := Stringify([]byte("bytes")); got != "bytes" {
		t.Fatalf("Stringify []byte = %q", got)
	}
	if got := Stringify(SafeString("safe")); got != "safe" {
		t.Fatalf("Stringify SafeString = %q", got)
	}
	if got := Stringify(testStringer{}); got != "stringer" {
		t.Fatalf("Stringify Stringer = %q", got)
	}
	if got := Stringify(errors.New("err")); got != "err" {
		t.Fatalf("Stringify error = %q", got)
	}
}

func TestWriteEscaped(t *testing.T) {
	var b strings.Builder
	if err := WriteEscaped(&b, "<b>"); err != nil {
		t.Fatalf("WriteEscaped error: %v", err)
	}
	if got := b.String(); got != "&lt;b&gt;" {
		t.Fatalf("WriteEscaped output = %q", got)
	}

	b.Reset()
	if err := WriteEscaped(&b, SafeString("<i>")); err != nil {
		t.Fatalf("WriteEscaped SafeString error: %v", err)
	}
	if got := b.String(); got != "<i>" {
		t.Fatalf("WriteEscaped SafeString output = %q", got)
	}
}

func TestWriteRaw(t *testing.T) {
	var b strings.Builder
	if err := WriteRaw(&b, "<b>"); err != nil {
		t.Fatalf("WriteRaw error: %v", err)
	}
	if got := b.String(); got != "<b>" {
		t.Fatalf("WriteRaw output = %q", got)
	}
}

func TestWriteNil(t *testing.T) {
	if err := WriteEscaped(nil, "value"); err != nil {
		t.Fatalf("WriteEscaped nil writer error: %v", err)
	}
	if err := WriteRaw(nil, "value"); err != nil {
		t.Fatalf("WriteRaw nil writer error: %v", err)
	}
}
