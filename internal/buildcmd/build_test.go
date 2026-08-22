package buildcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/internal/compiler"
)

func TestModuleFilePinsVersionAndGoVersion(t *testing.T) {
	got, err := moduleFile("v0.1.4", "")
	if err != nil {
		t.Fatalf("moduleFile failed: %v", err)
	}
	if !strings.Contains(got, "go "+requiredGoVersion+"\n") {
		t.Fatalf("module file does not require Go %s:\n%s", requiredGoVersion, got)
	}
	if !strings.Contains(got, "require "+modulePath+" v0.1.4") {
		t.Fatalf("module file does not pin go-hbars v0.1.4:\n%s", got)
	}
	if strings.Contains(got, "replace ") {
		t.Fatalf("unexpected replacement in release module:\n%s", got)
	}
}

func TestRequiredGoVersionMatchesModuleAndDocumentation(t *testing.T) {
	repoRoot := repositoryRoot(t)
	goMod, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "\ngo "+requiredGoVersion+"\n") {
		t.Fatalf("go.mod and generated renderer disagree on Go %s", requiredGoVersion)
	}
	readme, err := os.ReadFile(filepath.Join(repoRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "Go "+requiredGoVersion+" or later") {
		t.Fatalf("README does not document Go %s", requiredGoVersion)
	}
}

func TestResolveOptionsDefaults(t *testing.T) {
	root := t.TempDir()
	got, err := resolveOptions(Options{RootPath: root})
	if err != nil {
		t.Fatalf("resolveOptions failed: %v", err)
	}
	if got.TemplatesPath != filepath.Join(root, ".processor", "templates") {
		t.Errorf("templates path = %q", got.TemplatesPath)
	}
	if got.DataPath != "data" || got.SharedPath != "shared" || got.OutputPath != "pages" {
		t.Errorf("unexpected site path defaults: %#v", got)
	}
	wantVersion := "v" + strings.TrimPrefix(compiler.Version, "v")
	if got.ModuleVersion != wantVersion {
		t.Errorf("module version = %q, want %s", got.ModuleVersion, wantVersion)
	}
}

func TestRunCLIHelp(t *testing.T) {
	var output bytes.Buffer
	if err := RunCLI([]string{"-help"}, &output); err != nil {
		t.Fatalf("RunCLI help failed: %v", err)
	}
	if !strings.Contains(output.String(), "-templates-path") {
		t.Fatalf("help does not describe build flags:\n%s", output.String())
	}
}

func TestRunGeneratesSite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated renderer build in short mode")
	}

	repoRoot := repositoryRoot(t)
	root := t.TempDir()
	templatesDir := filepath.Join(root, ".processor", "templates")
	dataDir := filepath.Join(root, "data")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "greeting.hbs"), []byte("Hello {{title}}"), 0o644); err != nil {
		t.Fatal(err)
	}
	data := `{"_page":{"template":"greeting","output":"index.html"},"title":"Compiled World"}`
	if err := os.WriteFile(filepath.Join(dataDir, "index.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(root, "public")
	result, err := Run(Options{
		RootPath:      root,
		OutputPath:    outputPath,
		ModuleReplace: repoRoot,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.TemplateCompile <= 0 || result.GoBuild <= 0 || result.Render <= 0 {
		t.Fatalf("missing build timings: %#v", result)
	}
	content, err := os.ReadFile(filepath.Join(outputPath, "index.html"))
	if err != nil {
		t.Fatalf("read generated page: %v", err)
	}
	if string(content) != "Hello Compiled World" {
		t.Fatalf("generated page = %q", content)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
