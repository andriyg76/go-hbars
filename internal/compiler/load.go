package compiler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/andriyg76/hexerr"
)

// LoadTemplates loads all .hbs files from a directory recursively, or a single
// .hbs file. Template names use slash-separated paths without the extension.
func LoadTemplates(input string) (map[string]string, map[string]string, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, nil, err
	}
	templates := make(map[string]string)
	templateFiles := make(map[string]string)
	if info.IsDir() {
		err := filepath.WalkDir(input, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hbs") {
				return nil
			}

			rel, err := filepath.Rel(input, path)
			if err != nil {
				return hexerr.Wrapf(err, "resolve template name for %s", path)
			}
			name := filepath.ToSlash(strings.TrimSuffix(rel, ".hbs"))
			if name == "" || name == "." {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return hexerr.Wrapf(err, "read %s", path)
			}
			templates[name] = string(content)
			templateFiles[name] = path
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
		return templates, templateFiles, nil
	}
	if !strings.HasSuffix(input, ".hbs") {
		return nil, nil, hexerr.New("single file input must be .hbs")
	}
	content, err := os.ReadFile(input)
	if err != nil {
		return nil, nil, err
	}
	name := strings.TrimSuffix(filepath.Base(input), ".hbs")
	templates[name] = string(content)
	templateFiles[name] = input
	return templates, templateFiles, nil
}
