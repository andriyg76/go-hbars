package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSharedData(t *testing.T) {
	tmpDir := t.TempDir()
	sharedDir := filepath.Join(tmpDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}

	// Create a shared JSON file
	siteFile := filepath.Join(sharedDir, "site.json")
	if err := os.WriteFile(siteFile, []byte(`{"name": "Test Site", "url": "https://example.com"}`), 0644); err != nil {
		t.Fatalf("failed to create site.json: %v", err)
	}

	data, err := LoadSharedData(sharedDir)
	if err != nil {
		t.Fatalf("LoadSharedData error: %v", err)
	}

	site, ok := data["site"].(map[string]any)
	if !ok {
		t.Fatalf("expected site map, got %T", data["site"])
	}
	if site["name"] != "Test Site" {
		t.Errorf("expected name=Test Site, got %v", site["name"])
	}
}

func TestLoadSharedData_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	sharedDir := filepath.Join(tmpDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}

	data, err := LoadSharedData(sharedDir)
	if err != nil {
		t.Fatalf("LoadSharedData error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty map, got %v", data)
	}
}

func TestLoadSharedData_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	sharedDir := filepath.Join(tmpDir, "nonexistent")

	data, err := LoadSharedData(sharedDir)
	if err != nil {
		t.Fatalf("LoadSharedData error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty map for non-existent dir, got %v", data)
	}
}

func TestMergeSharedData(t *testing.T) {
	pageData := map[string]any{
		"title": "Page Title",
	}
	sharedData := map[string]any{
		"site": map[string]any{
			"name": "Test Site",
		},
	}

	MergeSharedData(pageData, sharedData)

	if pageData["title"] != "Page Title" {
		t.Error("page data should remain")
	}
	shared, ok := pageData["_shared"].(map[string]any)
	if !ok {
		t.Fatal("expected _shared key")
	}
	if shared["site"] == nil {
		t.Error("shared data should be merged")
	}
}

