package runtime

import (
	"io"
	"strings"
	"testing"
)

func TestLazyWriter_WriteAndFlush(t *testing.T) {
	var out strings.Builder
	lw := NewLazyWriter(&out)
	lw.Write([]byte("a"))
	lw.WriteLazyBlock(func(w io.Writer) {
		w.Write([]byte("[lazy]"))
	})
	lw.Write([]byte("b"))
	if err := lw.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	got := out.String()
	want := "a[lazy]b"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestLazyWriter_MissingPartialOutputAsLazy(t *testing.T) {
	var out strings.Builder
	lw := NewLazyWriter(&out)
	lw.Write([]byte("<div>"))
	MissingPartialOutput(lw, "")
	lw.Write([]byte("</div>"))
	if err := lw.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `partial "" is not defined`) {
		t.Errorf("output should contain partial error message, got %q", got)
	}
	if got != "<div><!-- partial \"\" is not defined --></div>" {
		t.Errorf("got %q", got)
	}
}

func TestMissingPartialOutput_PlainWriter(t *testing.T) {
	var out strings.Builder
	MissingPartialOutput(&out, "foo")
	got := out.String()
	want := `<!-- partial "foo" is not defined -->`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
