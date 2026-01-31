package sitegen

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

type testRenderer struct{}

func (testRenderer) Render(templateName string, w io.Writer, data any) error {
	_, _ = w.Write([]byte("<html>ok</html>"))
	return nil
}

func TestNewProcessor_NilConfig(t *testing.T) {
	p, err := NewProcessor(nil, testRenderer{})
	if err != nil {
		t.Fatalf("NewProcessor(nil): %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil Processor")
	}
	cfg := p.Config()
	if cfg.DataPath != "data" || cfg.OutputPath != "pages" {
		t.Errorf("default config: DataPath=%q OutputPath=%q", cfg.DataPath, cfg.OutputPath)
	}
}

func TestNewProcessor_WithRoot(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "out"}
	p, err := NewProcessor(cfg, testRenderer{})
	if err != nil {
		t.Fatalf("NewProcessor: %v", err)
	}
	if p.Config().RootPath != tmpDir {
		t.Errorf("Config().RootPath = %q, want %q", p.Config().RootPath, tmpDir)
	}
}

func TestProcessor_ProcessFile(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	sharedDir := filepath.Join(tmpDir, "shared")
	for _, d := range []string{dataDir, sharedDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	dataFile := filepath.Join(dataDir, "page.json")
	content := `{"_page":{"template":"main","output":"page.html"},"title":"Hi"}`
	if err := os.WriteFile(dataFile, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	p, err := NewProcessor(cfg, testRenderer{})
	if err != nil {
		t.Fatalf("NewProcessor: %v", err)
	}

	outPath, outContent, err := p.ProcessFile(dataFile)
	if err != nil {
		t.Fatalf("ProcessFile: %v", err)
	}
	if outPath == "" {
		t.Fatal("expected non-empty output path")
	}
	if len(outContent) == 0 {
		t.Error("expected non-empty content")
	}
	if string(outContent) != "<html>ok</html>" {
		t.Errorf("content = %q", outContent)
	}
}

func TestProcessor_Process(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	sharedDir := filepath.Join(tmpDir, "shared")
	for _, d := range []string{dataDir, sharedDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	pagePath := filepath.Join(dataDir, "index.json")
	if err := os.WriteFile(pagePath, []byte(`{"_page":{"template":"main","output":"index.html"},"x":1}`), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := &Config{RootPath: tmpDir, DataPath: "data", SharedPath: "shared", OutputPath: "pages"}
	p, err := NewProcessor(cfg, testRenderer{})
	if err != nil {
		t.Fatalf("NewProcessor: %v", err)
	}
	if err := p.Process(); err != nil {
		t.Fatalf("Process: %v", err)
	}

	indexHTML := filepath.Join(tmpDir, "pages", "index.html")
	got, err := os.ReadFile(indexHTML)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty index.html")
	}
}
