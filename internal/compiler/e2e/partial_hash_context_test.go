package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/internal/compiler"
)

func TestE2E_PartialHashValueUsesCallerContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	code, err := compiler.CompileTemplates(map[string]string{
		"default": `{{> menu menu=_shared.menu_main_teatr}}`,
		"menu":    `{{#each menu.items}}{{caption}}|{{/each}}`,
	}, compiler.Options{PackageName: "templates", MaxPhase: compiler.Phase4})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	src := string(code)
	if !strings.Contains(src, `"menu": data.Shared().MenuMainTeatr(),`) {
		t.Fatalf("partial hash value was not resolved from caller context:\n%s", grepSnippet(src, `"menu":`))
	}

	tmpDir := t.TempDir()
	repoPath := strings.ReplaceAll(repoRoot(t), "\\", "/")
	writeFile := func(path, body string) {
		full := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeFile("go.mod", `module test-partial-hash

go 1.24

replace github.com/andriyg76/go-hbars => `+repoPath+`
`)
	writeFile("templates/templates_gen.go", src)
	writeFile("main.go", `package main

import (
	"fmt"
	"os"

	templates "test-partial-hash/templates"
)

func main() {
	data := map[string]any{
		"_shared": map[string]any{
			"menu_main_teatr": map[string]any{
				"items": []any{
					map[string]any{"caption": "Home"},
					map[string]any{"caption": "History"},
				},
			},
		},
	}
	out, err := templates.RenderDefaultString(templates.DefaultContextFromMap(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if out != "Home|History|" {
		fmt.Fprintf(os.Stderr, "unexpected output: %q\n", out)
		os.Exit(1)
	}
}
`)

	cmd := exec.Command("go", "run", "-mod=mod", ".")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated renderer failed: %v\n%s", err, output)
	}
}
