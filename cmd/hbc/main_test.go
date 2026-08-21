package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplatesRecursively(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mustWrite("page.hbs", "page")
	mustWrite("news/item.hbs", "item")
	mustWrite("parts/header.hbs", "header")
	mustWrite("parts/ignored.txt", "ignored")

	templates, files, err := loadTemplates(root)
	if err != nil {
		t.Fatalf("loadTemplates: %v", err)
	}

	want := map[string]string{
		"page":         "page",
		"news/item":    "item",
		"parts/header": "header",
	}
	if len(templates) != len(want) {
		t.Fatalf("loaded %d templates, want %d: %#v", len(templates), len(want), templates)
	}
	for name, content := range want {
		if templates[name] != content {
			t.Errorf("template %q = %q, want %q", name, templates[name], content)
		}
		if files[name] == "" {
			t.Errorf("template file path missing for %q", name)
		}
	}
}
