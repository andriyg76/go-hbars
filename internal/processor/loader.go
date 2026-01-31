package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andriyg76/hexerr"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// LoadDataFile loads and parses a data file (JSON, YAML, or TOML).
func LoadDataFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, hexerr.Wrapf(err, "failed to read file %q", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var result map[string]any

	switch ext {
	case ".json", ".json5":
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, hexerr.Wrapf(err, "failed to parse JSON file %q", path)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, hexerr.Wrapf(err, "failed to parse YAML file %q", path)
		}
	case ".toml":
		if err := toml.Unmarshal(data, &result); err != nil {
			return nil, hexerr.Wrapf(err, "failed to parse TOML file %q", path)
		}
	default:
		return nil, hexerr.New(fmt.Sprintf("unsupported file format: %q (supported: .json, .yaml, .yml, .toml)", ext))
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
		return nil, hexerr.New("_page section must be an object")
	}

	config := &PageConfig{}

	if template, ok := pageMap["template"].(string); ok {
		config.Template = template
	} else {
		return nil, hexerr.New("_page.template is required and must be a string")
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
