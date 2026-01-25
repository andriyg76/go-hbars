package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andriyg76/go-hbars/internal/compiler"
)

func main() {
	var inPath string
	var outPath string
	var pkgName string
	var runtimeImport string
	var extList string

	flag.StringVar(&inPath, "in", "", "input template file or directory")
	flag.StringVar(&outPath, "out", "templates_gen.go", "output Go file path")
	flag.StringVar(&pkgName, "pkg", "", "package name for generated code")
	flag.StringVar(&runtimeImport, "runtime-import", "", "override runtime import path")
	flag.StringVar(&extList, "ext", ".hbs,.handlebars", "comma-separated template extensions")
	flag.Parse()

	if inPath == "" {
		fatal(errors.New("missing -in path"))
	}
	if pkgName == "" {
		pkgName = defaultPackage(outPath)
	}

	exts := parseExts(extList)
	templates, err := loadTemplates(inPath, exts)
	if err != nil {
		fatal(err)
	}
	if len(templates) == 0 {
		fatal(fmt.Errorf("no templates found under %q", inPath))
	}

	code, err := compiler.CompileTemplates(templates, compiler.Options{
		PackageName:   pkgName,
		RuntimeImport: runtimeImport,
	})
	if err != nil {
		fatal(err)
	}

	if err := writeOutput(outPath, code); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "hbc:", err)
	os.Exit(1)
}

func parseExts(input string) map[string]bool {
	exts := make(map[string]bool)
	for _, raw := range strings.Split(input, ",") {
		ext := strings.ToLower(strings.TrimSpace(raw))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		exts[ext] = true
	}
	return exts
}

func loadTemplates(path string, exts map[string]bool) (map[string]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		name := templateName(path)
		return map[string]string{name: string(content)}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	templates := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if len(exts) > 0 && !exts[ext] {
			continue
		}
		full := filepath.Join(path, entry.Name())
		content, err := os.ReadFile(full)
		if err != nil {
			return nil, err
		}
		name := templateName(full)
		if _, exists := templates[name]; exists {
			return nil, fmt.Errorf("duplicate template name %q", name)
		}
		templates[name] = string(content)
		names = append(names, name)
	}
	sort.Strings(names)
	return templates, nil
}

func templateName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func defaultPackage(outPath string) string {
	dir := filepath.Dir(outPath)
	if dir == "." || dir == "" {
		return "templates"
	}
	return filepath.Base(dir)
}

func writeOutput(outPath string, code []byte) error {
	dir := filepath.Dir(outPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(outPath, code, 0o644)
}
