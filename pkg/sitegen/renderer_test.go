package sitegen

import (
	"bytes"
	"io"
	"testing"
)

func TestLoadRenderer_Nil(t *testing.T) {
	_, err := LoadRenderer(nil)
	if err == nil {
		t.Fatal("LoadRenderer(nil) expected error")
	}
	if err.Error() == "" {
		t.Error("error message should be non-empty")
	}
}

func TestNewRendererFromFunctions(t *testing.T) {
	funcs := map[string]func(io.Writer, any) error{
		"main": func(w io.Writer, data any) error {
			_, _ = w.Write([]byte("main"))
			return nil
		},
	}
	r := NewRendererFromFunctions(funcs)
	if r == nil {
		t.Fatal("NewRendererFromFunctions returned nil")
	}
	var buf bytes.Buffer
	if err := r.Render("main", &buf, nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if buf.String() != "main" {
		t.Errorf("got %q, want main", buf.String())
	}
}

func TestAutoLoadRenderer_Map(t *testing.T) {
	funcMap := map[string]RenderFunc{
		"main": func(w io.Writer, data any) error {
			_, _ = w.Write([]byte("ok"))
			return nil
		},
	}
	r, err := AutoLoadRenderer(funcMap)
	if err != nil {
		t.Fatalf("AutoLoadRenderer: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
	var buf bytes.Buffer
	if err := r.Render("main", &buf, nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if buf.String() != "ok" {
		t.Errorf("got %q, want ok", buf.String())
	}
}

func TestLoadRendererFromPackage_FuncMap(t *testing.T) {
	funcs := map[string]func(io.Writer, any) error{
		"main": func(w io.Writer, data any) error {
			_, _ = w.Write([]byte("pkg"))
			return nil
		},
	}
	r, err := LoadRendererFromPackage(funcs)
	if err != nil {
		t.Fatalf("LoadRendererFromPackage: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
	var buf bytes.Buffer
	if err := r.Render("main", &buf, nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if buf.String() != "pkg" {
		t.Errorf("got %q, want pkg", buf.String())
	}
}
