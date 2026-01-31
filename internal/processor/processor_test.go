package processor

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.DataPath != "data" {
		t.Errorf("DataPath = %q, want data", c.DataPath)
	}
	if c.SharedPath != "shared" {
		t.Errorf("SharedPath = %q, want shared", c.SharedPath)
	}
	if c.OutputPath != "pages" {
		t.Errorf("OutputPath = %q, want pages", c.OutputPath)
	}
}

func TestNewProcessor(t *testing.T) {
	cfg := &Config{DataPath: "d", OutputPath: "o"}
	r := &mockRenderer{name: "main"}
	p := NewProcessor(cfg, r)
	if p == nil {
		t.Fatal("NewProcessor returned nil")
	}
	if p.Config() != cfg {
		t.Error("Config() did not return same config")
	}
}

func TestProcessor_ProcessFile(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dataFile := filepath.Join(dataDir, "page.json")
	dataContent := `{"_page": {"template": "main", "output": "out.html"}, "title": "Hi"}`
	if err := os.WriteFile(dataFile, []byte(dataContent), 0644); err != nil {
		t.Fatalf("write data: %v", err)
	}

	sharedDir := filepath.Join(tmpDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("mkdir shared: %v", err)
	}

	cfg := &Config{
		RootPath:   tmpDir,
		DataPath:   "data",
		SharedPath: "shared",
		OutputPath: "pages",
	}
	mock := &mockRenderer{name: "main"}
	p := NewProcessor(cfg, mock)

	sharedData := map[string]any{}
	outputPath, content, err := p.ProcessFile(dataFile, sharedData)
	if err != nil {
		t.Fatalf("ProcessFile: %v", err)
	}
	if outputPath == "" {
		t.Fatal("expected non-empty output path")
	}
	if !strings.HasSuffix(outputPath, "out.html") {
		t.Errorf("output path = %q, want suffix out.html", outputPath)
	}
	if !filepath.IsAbs(outputPath) {
		// resolvePath with RootPath should give path under tmpDir
		expectedPrefix := filepath.Join(tmpDir, "pages")
		if !strings.HasPrefix(filepath.Clean(outputPath), filepath.Clean(expectedPrefix)) {
			t.Errorf("output path = %q, expected under %q", outputPath, expectedPrefix)
		}
	}
	if len(content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestProcessor_ProcessFile_NoPageConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	dataFile := filepath.Join(dataDir, "no_page.json")
	if err := os.WriteFile(dataFile, []byte(`{"title": "No _page"}`), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	p := NewProcessor(cfg, &mockRenderer{})

	outputPath, content, err := p.ProcessFile(dataFile, nil)
	if err != nil {
		t.Fatalf("ProcessFile: %v", err)
	}
	if outputPath != "" || content != nil {
		t.Errorf("expected nil output for missing _page; got path=%q content=%v", outputPath, content != nil)
	}
}

func TestProcessor_Process_Integration(t *testing.T) {
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

	cfg := &Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	p := NewProcessor(cfg, &mockRenderer{name: "main"})

	if err := p.Process(); err != nil {
		t.Fatalf("Process: %v", err)
	}

	indexPath := filepath.Join(tmpDir, "pages", "index.html")
	got, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty index.html")
	}
}

type mockRenderer struct {
	name string
	out  io.Writer
}

func (m *mockRenderer) Render(templateName string, w io.Writer, data any) error {
	target := w
	if m.out != nil {
		target = m.out
	}
	_, _ = target.Write([]byte("<html>mock</html>"))
	return nil
}
