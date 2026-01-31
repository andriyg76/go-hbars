package sitegen

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type serverTestRenderer struct{}

func (serverTestRenderer) Render(templateName string, w io.Writer, data any) error {
	_, _ = w.Write([]byte("<html>ok</html>"))
	return nil
}

func TestNewServer_NilConfig(t *testing.T) {
	srv, err := NewServer(nil, serverTestRenderer{})
	if err != nil {
		t.Fatalf("NewServer(nil): %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil Server")
	}
	if srv.Config() == nil {
		t.Fatal("Config() should not be nil")
	}
	if srv.Config().Addr != ":8080" {
		t.Errorf("default Addr = %q, want :8080", srv.Config().Addr)
	}
	if srv.Address() != ":8080" {
		t.Errorf("Address() = %q, want :8080", srv.Address())
	}
}

func TestNewServer_WithRoot(t *testing.T) {
	tmpDir := t.TempDir()
	sharedDir := filepath.Join(tmpDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg := &Config{RootPath: tmpDir, Addr: ":0", DataPath: "data", SharedPath: "shared"}
	srv, err := NewServer(cfg, serverTestRenderer{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if srv.Config().RootPath != tmpDir {
		t.Errorf("Config().RootPath = %q", srv.Config().RootPath)
	}
	if srv.Address() != ":0" {
		t.Errorf("Address() = %q", srv.Address())
	}
}

func TestServer_Shutdown(t *testing.T) {
	tmpDir := t.TempDir()
	sharedDir := filepath.Join(tmpDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg := &Config{RootPath: tmpDir, Addr: ":0", DataPath: "data", SharedPath: "shared"}
	srv, err := NewServer(cfg, serverTestRenderer{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	// Shutdown without Start is a no-op; should not panic
	if err := srv.Shutdown(); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

func TestServer_Config(t *testing.T) {
	cfg := &Config{Addr: ":9999", DataPath: "d", SharedPath: "s"}
	srv, err := NewServer(cfg, serverTestRenderer{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	got := srv.Config()
	if got.Addr != ":9999" || got.DataPath != "d" || got.SharedPath != "s" {
		t.Errorf("Config() = %+v", got)
	}
}

func TestServer_HandlerServesRequest(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	sharedDir := filepath.Join(tmpDir, "shared")
	for _, d := range []string{dataDir, sharedDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(dataDir, "index.json"), []byte(`{"_page":{"template":"main","output":"index.html"},"x":1}`), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &Config{RootPath: tmpDir, Addr: ":0", DataPath: "data", SharedPath: "shared"}
	srv, err := NewServer(cfg, serverTestRenderer{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	req := newRequest(http.MethodGet, "/", nil)
	rec := newRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.code)
	}
	if rec.body != "<html>ok</html>" {
		t.Errorf("body = %q", rec.body)
	}
}

// minimal response recorder and request for testing handler without starting server
type responseRecorder struct {
	code int
	body string
}

func newRecorder() *responseRecorder {
	return &responseRecorder{code: 200}
}

func (r *responseRecorder) Header() http.Header       { return http.Header{} }
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = string(b)
	return len(b), nil
}
func (r *responseRecorder) WriteHeader(code int) { r.code = code }

func newRequest(method, path string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, path, body)
	return req
}
