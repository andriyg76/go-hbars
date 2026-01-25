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
	var helperFlags helperFlag

	flag.StringVar(&inPath, "in", "", "input template file or directory")
	flag.StringVar(&outPath, "out", "templates_gen.go", "output Go file path")
	flag.StringVar(&pkgName, "pkg", "", "package name for generated code")
	flag.StringVar(&runtimeImport, "runtime-import", "", "override runtime import path")
	flag.StringVar(&extList, "ext", ".hbs,.handlebars", "comma-separated template extensions")
	flag.Var(&helperFlags, "helper", "helper mapping name=Ident or name=import/path:Ident")
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
	helpers, err := parseHelpers(helperFlags)
	if err != nil {
		fatal(err)
	}

	code, err := compiler.CompileTemplates(templates, compiler.Options{
		PackageName:   pkgName,
		RuntimeImport: runtimeImport,
		Helpers:       helpers,
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

type helperFlag []string

func (h *helperFlag) String() string {
	if h == nil {
		return ""
	}
	return strings.Join(*h, ",")
}

func (h *helperFlag) Set(value string) error {
	if h == nil {
		return nil
	}
	*h = append(*h, value)
	return nil
}

func parseHelpers(values []string) (map[string]compiler.HelperRef, error) {
	helpers := make(map[string]compiler.HelperRef)
	for _, raw := range values {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid helper mapping %q", raw)
		}
		name := strings.TrimSpace(parts[0])
		ref := strings.TrimSpace(parts[1])
		if name == "" || ref == "" {
			return nil, fmt.Errorf("invalid helper mapping %q", raw)
		}
		var importPath, ident string
		if strings.Contains(ref, ":") {
			refParts := strings.SplitN(ref, ":", 2)
			importPath = strings.TrimSpace(refParts[0])
			ident = strings.TrimSpace(refParts[1])
			if importPath == "" || ident == "" {
				return nil, fmt.Errorf("invalid helper mapping %q", raw)
			}
		} else {
			ident = ref
		}
		if _, exists := helpers[name]; exists {
			return nil, fmt.Errorf("duplicate helper mapping for %q", name)
		}
		helpers[name] = compiler.HelperRef{ImportPath: importPath, Ident: ident}
	}
	return helpers, nil
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
