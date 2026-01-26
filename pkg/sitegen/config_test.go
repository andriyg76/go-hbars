package sitegen

import "testing"

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DataPath != "data" {
		t.Errorf("expected DataPath=data, got %q", config.DataPath)
	}
	if config.SharedPath != "shared" {
		t.Errorf("expected SharedPath=shared, got %q", config.SharedPath)
	}
	if config.TemplatesPath != ".processor/templates" {
		t.Errorf("expected TemplatesPath=.processor/templates, got %q", config.TemplatesPath)
	}
	if config.OutputPath != "pages" {
		t.Errorf("expected OutputPath=pages, got %q", config.OutputPath)
	}
	if config.Addr != ":8080" {
		t.Errorf("expected Addr=:8080, got %q", config.Addr)
	}
}

