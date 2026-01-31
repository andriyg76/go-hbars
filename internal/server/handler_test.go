package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/internal/processor"
)

type testRenderer struct{}

func (testRenderer) Render(templateName string, w io.Writer, data any) error {
	_, _ = w.Write([]byte("<html>ok</html>"))
	return nil
}

func TestNewHandler(t *testing.T) {
	cfg := &processor.Config{DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	proc := processor.NewProcessor(cfg, testRenderer{})
	h := NewHandler(proc, nil, "")
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestHandler_ServeHTTP_DataFile(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dataFile := filepath.Join(dataDir, "index.json")
	content := `{"_page":{"template":"main","output":"index.html"},"title":"Test"}`
	if err := os.WriteFile(dataFile, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &processor.Config{
		RootPath:   tmpDir,
		DataPath:   "data",
		SharedPath: "shared",
		OutputPath: "pages",
	}
	proc := processor.NewProcessor(cfg, testRenderer{})
	h := NewHandler(proc, map[string]any{}, "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Errorf("body = %q", body)
	}
}

func TestHandler_ServeHTTP_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &processor.Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	proc := processor.NewProcessor(cfg, testRenderer{})
	h := NewHandler(proc, nil, "")

	req := httptest.NewRequest(http.MethodGet, "/nonexistent-page", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandler_ServeHTTP_Static(t *testing.T) {
	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "style.css"), []byte("body{}"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &processor.Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	proc := processor.NewProcessor(cfg, testRenderer{})
	h := NewHandler(proc, nil, filepath.Join(tmpDir, "static"))

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "body") {
		t.Errorf("body = %q", rec.Body.String())
	}
}
