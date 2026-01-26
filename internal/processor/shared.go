package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadSharedData recursively loads shared data files from a directory.
// File names (without extension) become keys, and nested directories become nested objects.
func LoadSharedData(dirPath string) (map[string]any, error) {
	if dirPath == "" {
		return make(map[string]any), nil
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("failed to stat shared directory %q: %w", dirPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("shared path %q is not a directory", dirPath)
	}

	result := make(map[string]any)
	if err := loadSharedRecursive(dirPath, dirPath, result); err != nil {
		return nil, err
	}

	return result, nil
}

func loadSharedRecursive(basePath, currentPath string, result map[string]any) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %q: %w", currentPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Recursively process subdirectories
			subPath := filepath.Join(currentPath, entry.Name())
			subMap := make(map[string]any)
			if err := loadSharedRecursive(basePath, subPath, subMap); err != nil {
				return err
			}
			if len(subMap) > 0 {
				result[entry.Name()] = subMap
			}
			continue
		}

		// Process files
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".json5" && ext != ".yaml" && ext != ".yml" && ext != ".toml" {
			continue
		}

		filePath := filepath.Join(currentPath, entry.Name())
		data, err := LoadDataFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load shared file %q: %w", filePath, err)
		}

		// Remove _page section if present (shared files shouldn't have it)
		RemovePageConfig(data)

		// Use base name (without extension) as key
		baseName := strings.TrimSuffix(entry.Name(), ext)
		baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName)) // Handle .json5

		// If it's a single value object, unwrap it
		if len(data) == 1 {
			for k, v := range data {
				result[baseName] = map[string]any{k: v}
				break
			}
		} else {
			result[baseName] = data
		}
	}

	return nil
}

// MergeSharedData merges shared data into page data under the _shared key.
func MergeSharedData(pageData map[string]any, sharedData map[string]any) {
	if len(sharedData) > 0 {
		pageData["_shared"] = sharedData
	}
}

