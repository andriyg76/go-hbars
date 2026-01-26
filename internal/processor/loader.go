package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// LoadDataFile loads and parses a data file (JSON, YAML, or TOML).
func LoadDataFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var result map[string]any

	switch ext {
	case ".json", ".json5":
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON file %q: %w", path, err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse YAML file %q: %w", path, err)
		}
	case ".toml":
		if err := toml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse TOML file %q: %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %q (supported: .json, .yaml, .yml, .toml)", ext)
	}

	return result, nil
}

// PageConfig represents the _page configuration section.
type PageConfig struct {
	Template string `json:"template" yaml:"template" toml:"template"`
	Output   string `json:"output,omitempty" yaml:"output,omitempty" toml:"output,omitempty"`
}

// ExtractPageConfig extracts the _page configuration from data.
func ExtractPageConfig(data map[string]any) (*PageConfig, error) {
	pageRaw, ok := data["_page"]
	if !ok {
		return nil, nil // No _page section, file should be ignored
	}

	// Convert to map for easier handling
	pageMap, ok := pageRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("_page section must be an object")
	}

	config := &PageConfig{}

	if template, ok := pageMap["template"].(string); ok {
		config.Template = template
	} else {
		return nil, fmt.Errorf("_page.template is required and must be a string")
	}

	if output, ok := pageMap["output"].(string); ok {
		config.Output = output
	}

	return config, nil
}

// RemovePageConfig removes the _page section from data.
func RemovePageConfig(data map[string]any) {
	delete(data, "_page")
}

