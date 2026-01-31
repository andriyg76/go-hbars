package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

// TestE2E_Showcase_NilContext compiles showcase templates and renders with nil context.
// Ensures no panic and partials receive empty context via FromMap.
func TestE2E_Showcase_NilContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	root := repoRoot(t)
	tmplDir := filepath.Join(root, "examples", "showcase", "templates")
	tmpls := make(map[string]string)

	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		t.Fatalf("read templates dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hbs") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".hbs")
		content, err := os.ReadFile(filepath.Join(tmplDir, entry.Name()))
		if err != nil {
			t.Fatalf("read template %s: %v", entry.Name(), err)
		}
		tmpls[name] = string(content)
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
		t.Fatalf("compile: %v", err)
	}

	tmpDir := t.TempDir()
	repoPath := strings.ReplaceAll(root, "\\", "/")

	writeFile := func(path, body string) {
		full := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile("go.mod", `module test-showcase-nil

go 1.24

replace github.com/andriyg76/go-hbars => `+repoPath+`
`)
	writeFile("templates/templates_gen.go", string(code))
	writeFile("main.go", `package main

import (
	"fmt"
	"os"
	"strings"

	templates "test-showcase-nil/templates"
)

func main() {
	// Nil context: FromMap(nil) uses empty map; must not panic.
	// Dynamic partial (lookup . "cardPartial") with empty name: output goes to HTML, log.Error; no error returned.
	out, err := templates.RenderMainString(templates.MainContextFromMap(nil))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render nil: %v\n", err)
		os.Exit(1)
	}
	// Empty map context: same behaviour
	out2, err := templates.RenderMainString(templates.MainContextFromMap(map[string]any{}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render empty map: %v\n", err)
		os.Exit(1)
	}
	// Both should produce same (empty) structure
	s1 := strings.TrimSpace(out)
	s2 := strings.TrimSpace(out2)
	if s1 != s2 {
		fmt.Fprintf(os.Stderr, "nil and empty map output differ\n")
		os.Exit(1)
	}
	// Dynamic partial with empty name: error is in HTML (comment) and log, not as returned error
	if !strings.Contains(out, "partial \"\" is not defined") {
		fmt.Fprintf(os.Stderr, "expected partial error message in HTML output\n")
		os.Exit(1)
	}
	fmt.Println("OK")
}
`)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
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
