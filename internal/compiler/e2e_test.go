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
	"regexp"
	"strings"
	
	templates "test-templates/templates"
)

func normalizeWhitespace(s string) string {
	// Normalize whitespace for content comparison:
	// - Normalize spaces within lines (collapse multiple spaces, trim trailing)
	// - Normalize spaces around colons
	// - Collapse multiple newlines to single newline (preserve line structure but ignore extra blank lines)
	
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		// Trim trailing whitespace
		line = strings.TrimRight(line, " \t")
		// Collapse multiple spaces to single space
		line = regexp.MustCompile(" +").ReplaceAllString(line, " ")
		// Normalize colon spacing (": " or ":" -> ": ")
		line = regexp.MustCompile(":([^\\s])").ReplaceAllString(line, ": $1")
		lines[i] = line
	}
	s = strings.Join(lines, "\n")
	
	// Collapse multiple newlines to single newline (ignore extra blank lines)
	s = regexp.MustCompile("\\n{2,}").ReplaceAllString(s, "\n")
	
	// Trim leading and trailing newlines
	s = strings.Trim(s, "\n")
	
	return s
}

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
	
	// Normalize whitespace: collapse multiple consecutive newlines to single newline,
	// normalize spaces around colons, and trim trailing whitespace from lines
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

func TestE2E_IncludeZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	templates := map[string]string{
		"main": "{{#if count includeZero=true}}zero{{else}}nope{{/if}}",
	}
	data := map[string]any{"count": 0}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	code, err := CompileTemplates(templates, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	tmpDir := t.TempDir()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	writeFile := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile(filepath.Join(tmpDir, "go.mod"), `module test-includezero

go 1.24

replace github.com/andriyg76/go-hbars => `+strings.ReplaceAll(repoRoot, "\\", "/")+`
`)
	if err := os.MkdirAll(filepath.Join(tmpDir, "templates"), 0755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	writeFile(filepath.Join(tmpDir, "templates", "templates_gen.go"), string(code))
	writeFile(filepath.Join(tmpDir, "data.json"), string(dataBytes))
	writeFile(filepath.Join(tmpDir, "main.go"), `package main

import (
	"encoding/json"
	"fmt"
	"os"
	templates "test-includezero/templates"
)

func main() {
	dataBytes, _ := os.ReadFile("data.json")
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "json: %v\n", err)
		os.Exit(1)
	}
	out, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	out = strings.TrimSpace(out)
	if out != "zero" {
		fmt.Fprintf(os.Stderr, "got %q want \"zero\"\n", out)
		os.Exit(1)
	}
	fmt.Println("OK")
}
`)
	// Fix main.go: add strings import
	mainContent, _ := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	mainStr := string(mainContent)
	mainStr = strings.Replace(mainStr, `import (
	"encoding/json"
	"fmt"
	"os"
	templates "test-includezero/templates"
)`, `import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	templates "test-includezero/templates"
)`, 1)
	writeFile(filepath.Join(tmpDir, "main.go"), mainStr)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("go mod tidy: %v", err)
	}
	cmd = exec.Command("go", "run", ".")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run: %v\nOutput:\n%s", err, output)
	}
	if !strings.Contains(string(output), "OK") {
		t.Fatalf("expected OK in output, got:\n%s", output)
	}
}

func TestE2E_LayoutBlocksDirectionB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	// Direction B (layout-first, lazy slots): layout runs first, content runs when layout does {{> main}}
	templates := map[string]string{
		"layout": `<head>{{#block "header"}}Default{{/block}}</head><body>{{> main}}</body>`,
		"main":   `{{#partial "header"}}Custom Header{{/partial}}body content`,
	}
	data := map[string]any{}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	code, err := CompileTemplates(templates, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"block":   {ImportPath: "github.com/andriyg76/go-hbars/runtime", Ident: "Block"},
			"partial": {ImportPath: "github.com/andriyg76/go-hbars/runtime", Ident: "Partial"},
		},
		LayoutContent: &LayoutContentConfig{Layout: "layout", Content: "main"},
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	tmpDir := t.TempDir()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	writeFile := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile(filepath.Join(tmpDir, "go.mod"), `module test-layout-b

go 1.24

replace github.com/andriyg76/go-hbars => `+strings.ReplaceAll(repoRoot, "\\", "/")+`
`)
	if err := os.MkdirAll(filepath.Join(tmpDir, "templates"), 0755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	writeFile(filepath.Join(tmpDir, "templates", "templates_gen.go"), string(code))
	writeFile(filepath.Join(tmpDir, "data.json"), string(dataBytes))
	writeFile(filepath.Join(tmpDir, "main.go"), `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	templates "test-layout-b/templates"
)

func main() {
	dataBytes, _ := os.ReadFile("data.json")
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "json: %v\n", err)
		os.Exit(1)
	}
	out, err := templates.RenderWithLayoutString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	out = strings.TrimSpace(out)
	want := "<head>Custom Header</head><body>body content</body>"
	if out != want {
		fmt.Fprintf(os.Stderr, "got %q want %q\n", out, want)
		os.Exit(1)
	}
	fmt.Println("OK")
}
`)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("go mod tidy: %v", err)
	}
	cmd = exec.Command("go", "run", ".")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run: %v\nOutput:\n%s", err, output)
	}
	if !strings.Contains(string(output), "OK") {
		t.Fatalf("expected OK in output, got:\n%s", output)
	}
}

func TestE2E_LayoutBlocks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}
	// Direction A: page defines slots with {{#partial}}, then invokes layout; layout uses {{#block}}
	templates := map[string]string{
		"main":   `{{#partial "header"}}Custom Header{{/partial}}{{> layout}}`,
		"layout": `<head>{{#block "header"}}Default{{/block}}</head><body>ok</body>`,
	}
	data := map[string]any{}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	// Only block/partial needed for layout-blocks; full registry would add unused handlebars import
	code, err := CompileTemplates(templates, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"block":   {ImportPath: "github.com/andriyg76/go-hbars/runtime", Ident: "Block"},
			"partial": {ImportPath: "github.com/andriyg76/go-hbars/runtime", Ident: "Partial"},
		},
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	tmpDir := t.TempDir()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	writeFile := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile(filepath.Join(tmpDir, "go.mod"), `module test-layoutblocks

go 1.24

replace github.com/andriyg76/go-hbars => `+strings.ReplaceAll(repoRoot, "\\", "/")+`
`)
	if err := os.MkdirAll(filepath.Join(tmpDir, "templates"), 0755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	writeFile(filepath.Join(tmpDir, "templates", "templates_gen.go"), string(code))
	writeFile(filepath.Join(tmpDir, "data.json"), string(dataBytes))
	writeFile(filepath.Join(tmpDir, "main.go"), `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	templates "test-layoutblocks/templates"
)

func main() {
	dataBytes, _ := os.ReadFile("data.json")
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "json: %v\n", err)
		os.Exit(1)
	}
	out, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	out = strings.TrimSpace(out)
	want := "<head>Custom Header</head><body>ok</body>"
	if out != want {
		fmt.Fprintf(os.Stderr, "got %q want %q\n", out, want)
		os.Exit(1)
	}
	fmt.Println("OK")
}
`)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("go mod tidy: %v", err)
	}
	cmd = exec.Command("go", "run", ".")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run: %v\nOutput:\n%s", err, output)
	}
	if !strings.Contains(string(output), "OK") {
		t.Fatalf("expected OK in output, got:\n%s", output)
	}
}

