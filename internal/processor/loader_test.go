package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDataFile_JSON(t *testing.T) {
	// Create a temporary JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(jsonFile, []byte(`{"name": "test", "value": 42}`), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	data, err := LoadDataFile(jsonFile)
	if err != nil {
		t.Fatalf("LoadDataFile error: %v", err)
	}

	if data["name"] != "test" {
		t.Errorf("expected name=test, got %v", data["name"])
	}
	if data["value"] != float64(42) {
		t.Errorf("expected value=42, got %v", data["value"])
	}
}

func TestExtractPageConfig(t *testing.T) {
	data := map[string]any{
		"_page": map[string]any{
			"template": "main",
			"output":   "index.html",
		},
		"title": "Test",
	}

	config, err := ExtractPageConfig(data)
	if err != nil {
		t.Fatalf("ExtractPageConfig error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Template != "main" {
		t.Errorf("expected template=main, got %q", config.Template)
	}
	if config.Output != "index.html" {
		t.Errorf("expected output=index.html, got %q", config.Output)
	}
}

func TestExtractPageConfig_Missing(t *testing.T) {
	data := map[string]any{
		"title": "Test",
	}

	config, err := ExtractPageConfig(data)
	if err != nil {
		t.Fatalf("ExtractPageConfig error: %v", err)
	}
	if config != nil {
		t.Fatal("expected nil config for missing _page")
	}
}

func TestRemovePageConfig(t *testing.T) {
	data := map[string]any{
		"_page": map[string]any{
			"template": "main",
		},
		"title": "Test",
	}

	RemovePageConfig(data)

	if _, ok := data["_page"]; ok {
		t.Error("_page should be removed")
	}
	if data["title"] != "Test" {
		t.Error("other fields should remain")
	}
}

