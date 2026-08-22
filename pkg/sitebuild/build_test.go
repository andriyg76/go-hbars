package sitebuild_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/pkg/sitebuild"
)

func TestPublicAPI(t *testing.T) {
	var run func(sitebuild.Options) (sitebuild.Result, error) = sitebuild.Run
	if run == nil {
		t.Fatal("sitebuild.Run is nil")
	}
}

func TestCompile(t *testing.T) {
	templates := t.TempDir()
	if err := os.WriteFile(filepath.Join(templates, "hello.hbs"), []byte("Hello {{name}}"), 0o644); err != nil {
		t.Fatal(err)
	}

	generated, err := sitebuild.Compile(sitebuild.CompileOptions{
		TemplatesPath:     templates,
		PackageName:       "templates",
		GenerateBootstrap: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "package templates") || !strings.Contains(string(generated), "NewRenderer") {
		t.Fatalf("generated source is missing package or bootstrap:\n%s", generated)
	}
}
