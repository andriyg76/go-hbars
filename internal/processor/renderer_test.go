package processor

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewCompiledTemplateRenderer_Nil(t *testing.T) {
	r, err := NewCompiledTemplateRenderer(nil)
	if err != nil {
		t.Fatalf("NewCompiledTemplateRenderer(nil): %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
	var buf bytes.Buffer
	err = r.Render("main", &buf, nil)
	if err == nil {
		t.Error("Render with no templates expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v", err)
	}
}

func TestNewCompiledTemplateRenderer_FuncMap(t *testing.T) {
	var wrote string
	funcs := map[string]func(io.Writer, any) error{
		"main": func(w io.Writer, data any) error {
			wrote = "main"
			_, _ = w.Write([]byte("main"))
			return nil
		},
		"blog/post": func(w io.Writer, data any) error {
			wrote = "blog/post"
			_, _ = w.Write([]byte("blogpost"))
			return nil
		},
	}
	r, err := NewCompiledTemplateRenderer(funcs)
	if err != nil {
		t.Fatalf("NewCompiledTemplateRenderer(funcs): %v", err)
	}

	var buf bytes.Buffer
	if err := r.Render("main", &buf, nil); err != nil {
		t.Fatalf("Render(main): %v", err)
	}
	if buf.String() != "main" {
		t.Errorf("Render(main) wrote %q, want main", buf.String())
	}
	if wrote != "main" {
		t.Errorf("wrote = %q", wrote)
	}

	buf.Reset()
	if err := r.Render("blog/post", &buf, nil); err != nil {
		t.Fatalf("Render(blog/post): %v", err)
	}
	if buf.String() != "blogpost" {
		t.Errorf("Render(blog/post) wrote %q, want blogpost", buf.String())
	}
}

func TestCompiledTemplateRenderer_RegisterRenderFunc(t *testing.T) {
	r, _ := NewCompiledTemplateRenderer(nil)
	r.RegisterRenderFunc("custom", func(w io.Writer, data any) error {
		_, _ = w.Write([]byte("custom"))
		return nil
	})
	var buf bytes.Buffer
	if err := r.Render("custom", &buf, nil); err != nil {
		t.Fatalf("Render(custom): %v", err)
	}
	if buf.String() != "custom" {
		t.Errorf("got %q, want custom", buf.String())
	}
}

func TestCompiledTemplateRenderer_FindTemplateName_Normalize(t *testing.T) {
	funcs := map[string]func(io.Writer, any) error{
		"main": func(w io.Writer, data any) error { return nil },
	}
	r, _ := NewCompiledTemplateRenderer(funcs)
	var buf bytes.Buffer
	// Leading/trailing slashes should be normalized
	if err := r.Render("/main/", &buf, nil); err != nil {
		t.Fatalf("Render(/main/): %v", err)
	}
}

func TestCompiledTemplateRenderer_UnknownTemplate(t *testing.T) {
	funcs := map[string]func(io.Writer, any) error{
		"main": func(w io.Writer, data any) error { return nil },
	}
	r, _ := NewCompiledTemplateRenderer(funcs)
	var buf bytes.Buffer
	err := r.Render("nonexistent", &buf, nil)
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "nonexistent") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v", err)
	}
}
