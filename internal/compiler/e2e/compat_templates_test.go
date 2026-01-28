package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

func TestE2E_CompatTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	root := repoRoot(t)
	tmplDir := filepath.Join(root, "examples", "compat", "templates")
	tmpls := make(map[string]string)

	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		t.Fatalf("failed to read templates directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".hbs") {
			name := strings.TrimSuffix(entry.Name(), ".hbs")
			content, err := os.ReadFile(filepath.Join(tmplDir, entry.Name()))
			if err != nil {
				t.Fatalf("failed to read template %s: %v", entry.Name(), err)
			}
			tmpls[name] = string(content)
		}
	}

	dataPath := filepath.Join(root, "examples", "compat", "data.json")
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("failed to read data.json: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		t.Fatalf("failed to parse data.json: %v", err)
	}

	expectedPath := filepath.Join(root, "examples", "compat", "expected.txt")
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected.txt: %v", err)
	}

	helperRegistry := helpers.Registry()
	compilerHelpers := make(map[string]compiler.HelperRef, len(helperRegistry))
	for name, ref := range helperRegistry {
		compilerHelpers[name] = compiler.HelperRef{
			ImportPath: ref.ImportPath,
			Ident:      ref.Ident,
		}
	}
	opts := compiler.Options{
		PackageName: "templates",
		Helpers:     compilerHelpers,
	}

	code, err := compiler.CompileTemplates(tmpls, opts)
	if err != nil {
		t.Fatalf("failed to compile templates: %v", err)
	}

	tmpDir := t.TempDir()
	repoRootPath := strings.ReplaceAll(root, "\\", "/")

	goMod := `module test-templates

go 1.24

replace github.com/andriyg76/go-hbars => ` + repoRootPath + `
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	if err := os.Mkdir(filepath.Join(tmpDir, "templates"), 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "templates", "templates_gen.go"), code, 0644); err != nil {
		t.Fatalf("failed to write generated code: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "data.json"), dataBytes, 0644); err != nil {
		t.Fatalf("failed to copy data.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "expected.txt"), expectedBytes, 0644); err != nil {
		t.Fatalf("failed to copy expected.txt: %v", err)
	}

	testFile := `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	templates "test-templates/templates"
)

func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, " \t")
		line = regexp.MustCompile(" +").ReplaceAllString(line, " ")
		line = regexp.MustCompile(":([^\\s])").ReplaceAllString(line, ": $1")
		lines[i] = line
	}
	s = strings.Join(lines, "\n")
	s = regexp.MustCompile("\\n{2,}").ReplaceAllString(s, "\n")
	return strings.Trim(s, "\n")
}

func main() {
	dataBytes, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read data: %v\n", err)
		os.Exit(1)
	}
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse data: %v\n", err)
		os.Exit(1)
	}
	output, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to render: %v\n", err)
		os.Exit(1)
	}
	expectedBytes, err := os.ReadFile("expected.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read expected: %v\n", err)
		os.Exit(1)
	}
	expected := string(expectedBytes)
	output = strings.ReplaceAll(output, "\r\n", "\n")
	expected = strings.ReplaceAll(expected, "\r\n", "\n")
	output = normalizeWhitespace(output)
	expected = normalizeWhitespace(expected)
	if output != expected {
		fmt.Fprintf(os.Stderr, "output mismatch!\n")
		fmt.Fprintf(os.Stderr, "Expected:\n%s\n", expected)
		fmt.Fprintf(os.Stderr, "Got:\n%s\n", output)
		os.Exit(1)
	}
	fmt.Println("OK")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(testFile), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run go mod tidy: %v", err)
	}

	cmd = exec.Command("go", "run", "main.go")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test failed: %v\nOutput:\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "OK") {
		t.Fatalf("unexpected output: %s", string(output))
	}
}
