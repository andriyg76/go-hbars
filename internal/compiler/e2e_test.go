package compiler

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/helpers"
)

func TestE2E_CompatTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Read templates
	templatesDir := filepath.Join("..", "..", "examples", "compat", "templates")
	templates := make(map[string]string)
	
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		t.Fatalf("failed to read templates directory: %v", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".hbs") {
			name := strings.TrimSuffix(entry.Name(), ".hbs")
			content, err := os.ReadFile(filepath.Join(templatesDir, entry.Name()))
			if err != nil {
				t.Fatalf("failed to read template %s: %v", entry.Name(), err)
			}
			templates[name] = string(content)
		}
	}
	
	// Read data
	dataPath := filepath.Join("..", "..", "examples", "compat", "data.json")
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("failed to read data.json: %v", err)
	}
	
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		t.Fatalf("failed to parse data.json: %v", err)
	}
	
	// Read expected output
	expectedPath := filepath.Join("..", "..", "examples", "compat", "expected.txt")
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected.txt: %v", err)
	}
	
	// Compile templates
	helperRegistry := helpers.Registry()
	compilerHelpers := make(map[string]HelperRef, len(helperRegistry))
	for name, ref := range helperRegistry {
		compilerHelpers[name] = HelperRef{
			ImportPath: ref.ImportPath,
			Ident:      ref.Ident,
		}
	}
	opts := Options{
		PackageName: "templates",
		Helpers:     compilerHelpers,
	}
	
	code, err := CompileTemplates(templates, opts)
	if err != nil {
		t.Fatalf("failed to compile templates: %v", err)
	}
	
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "go-hbars-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Get the repository root (assuming we're in internal/compiler)
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("failed to get repo root: %v", err)
	}
	
	// Write go.mod
	goMod := `module test-templates

go 1.24

replace github.com/andriyg76/go-hbars => ` + strings.ReplaceAll(repoRoot, "\\", "/") + `
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	
	// Create templates subdirectory
	tmpTemplatesDir := filepath.Join(tmpDir, "templates")
	if err := os.Mkdir(tmpTemplatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}
	
	// Write generated code
	if err := os.WriteFile(filepath.Join(tmpTemplatesDir, "templates_gen.go"), code, 0644); err != nil {
		t.Fatalf("failed to write generated code: %v", err)
	}
	
	// Copy data.json to temp directory
	if err := os.WriteFile(filepath.Join(tmpDir, "data.json"), dataBytes, 0644); err != nil {
		t.Fatalf("failed to copy data.json: %v", err)
	}
	
	// Copy expected.txt to temp directory
	if err := os.WriteFile(filepath.Join(tmpDir, "expected.txt"), expectedBytes, 0644); err != nil {
		t.Fatalf("failed to copy expected.txt: %v", err)
	}
	
	// Write test file
	testFile := `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	
	templates "test-templates/templates"
)

func main() {
	// Read data
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
	
	// Render template
	output, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to render: %v\n", err)
		os.Exit(1)
	}
	
	// Read expected
	expectedBytes, err := os.ReadFile("expected.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read expected: %v\n", err)
		os.Exit(1)
	}
	expected := string(expectedBytes)
	
	// Normalize line endings
	output = strings.ReplaceAll(output, "\r\n", "\n")
	expected = strings.ReplaceAll(expected, "\r\n", "\n")
	
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
	
	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run go mod tidy: %v", err)
	}
	
	// Run the test
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

