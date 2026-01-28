package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/internal/compiler"
)

func TestE2E_IncludeZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmpls := map[string]string{
		"main": "{{#if count includeZero=true}}zero{{else}}nope{{/if}}",
	}
	data := map[string]any{"count": 0}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	code, err := compiler.CompileTemplates(tmpls, compiler.Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	tmpDir := t.TempDir()
	root := repoRoot(t)
	repoRootPath := strings.ReplaceAll(root, "\\", "/")

	writeFile := func(path, content string) {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile(filepath.Join(tmpDir, "go.mod"), `module test-includezero

go 1.24

replace github.com/andriyg76/go-hbars => `+repoRootPath+`
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
