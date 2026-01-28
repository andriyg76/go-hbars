package compiler

import (
	"regexp"
	"strings"
	"testing"
)

func TestCompileTemplates_GenerateBootstrap(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main":   "Hello {{name}}",
		"header": "<header>{{title}}</header>",
	}, Options{
		PackageName:      "templates",
		GenerateBootstrap: true,
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)

	// Check for bootstrap imports (public packages only so user modules can use -bootstrap)
	if !strings.Contains(src, "github.com/andriyg76/go-hbars/pkg/renderer") {
		t.Fatalf("missing renderer import in bootstrap code")
	}
	if !strings.Contains(src, "github.com/andriyg76/go-hbars/pkg/sitegen") {
		t.Fatalf("missing sitegen import in bootstrap code")
	}

	// Check for rendererFuncs map
	if !strings.Contains(src, "var rendererFuncs = map[string]func(io.Writer, any) error") {
		t.Fatalf("missing rendererFuncs map")
	}
	if !regexp.MustCompile(`"main":\s+RenderMain`).MatchString(src) {
		t.Fatalf("missing main in rendererFuncs")
	}
	if !regexp.MustCompile(`"header":\s+RenderHeader`).MatchString(src) {
		t.Fatalf("missing header in rendererFuncs")
	}

	// Check for NewRenderer function
	if !strings.Contains(src, "func NewRenderer() renderer.TemplateRenderer") {
		t.Fatalf("missing NewRenderer function")
	}

	// Check for NewQuickProcessor function
	if !strings.Contains(src, "func NewQuickProcessor() (*sitegen.Processor, error)") {
		t.Fatalf("missing NewQuickProcessor function")
	}

	// Check for NewQuickServer function
	if !strings.Contains(src, "func NewQuickServer() (*sitegen.Server, error)") {
		t.Fatalf("missing NewQuickServer function")
	}
}

func TestCompileTemplates_NoBootstrap(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "Hello {{name}}",
	}, Options{
		PackageName:      "templates",
		GenerateBootstrap: false,
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)

	// Should not contain bootstrap code
	if strings.Contains(src, "NewRenderer") {
		t.Fatalf("unexpected NewRenderer in non-bootstrap code")
	}
	if strings.Contains(src, "NewQuickProcessor") {
		t.Fatalf("unexpected NewQuickProcessor in non-bootstrap code")
	}
	if strings.Contains(src, "NewQuickServer") {
		t.Fatalf("unexpected NewQuickServer in non-bootstrap code")
	}
}

