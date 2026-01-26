package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	helperspkg "github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

func main() {
	var inPath string
	var outPath string
	var pkgName string
	var runtimeImport string
	var extList string
	var helperFlags helperFlag
	var importFlags importFlag
	var helpersFlags helpersFlag
	var noCoreHelpers bool

	flag.StringVar(&inPath, "in", "", "input template file or directory")
	flag.StringVar(&outPath, "out", "templates_gen.go", "output Go file path")
	flag.StringVar(&pkgName, "pkg", "", "package name for generated code")
	flag.StringVar(&runtimeImport, "runtime-import", "", "override runtime import path")
	flag.StringVar(&extList, "ext", ".hbs,.handlebars", "comma-separated template extensions")
	flag.Var(&helperFlags, "helper", "helper mapping name=Ident or name=import/path:Ident (legacy)")
	flag.Var(&importFlags, "import", "import path for helpers: path or path:alias")
	flag.Var(&helpersFlags, "helpers", "comma-separated helper list: [alias:]Name or [alias:]name=Ident")
	flag.BoolVar(&noCoreHelpers, "no-core-helpers", false, "disable default core helpers registry")
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
	helpers, err := buildHelpers(noCoreHelpers, importFlags, helpersFlags, helperFlags)
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

type importFlag []string

func (i *importFlag) String() string {
	if i == nil {
		return ""
	}
	return strings.Join(*i, ",")
}

func (i *importFlag) Set(value string) error {
	if i == nil {
		return nil
	}
	*i = append(*i, value)
	return nil
}

type helpersFlag []string

func (h *helpersFlag) String() string {
	if h == nil {
		return ""
	}
	return strings.Join(*h, ",")
}

func (h *helpersFlag) Set(value string) error {
	if h == nil {
		return nil
	}
	*h = append(*h, value)
	return nil
}

// buildHelpers merges helpers from multiple sources with proper precedence:
// 1. Core helpers registry (unless -no-core-helpers)
// 2. -import/-helpers flags
// 3. Legacy -helper flags (highest precedence, can override)
func buildHelpers(noCoreHelpers bool, importFlags importFlag, helpersFlags helpersFlag, legacyHelperFlags helperFlag) (map[string]compiler.HelperRef, error) {
	helperMap := make(map[string]compiler.HelperRef)

	// Step 1: Add core helpers unless disabled
	if !noCoreHelpers {
		coreRegistry := helperspkg.Registry()
		for name, ref := range coreRegistry {
			helperMap[name] = compiler.HelperRef{
				ImportPath: ref.ImportPath,
				Ident:      ref.Ident,
			}
		}
	}

	// Step 2: Parse imports and build import map
	importMap, defaultImport, err := parseImports(importFlags)
	if err != nil {
		return nil, err
	}

	// Step 3: Parse -helpers flags and add to helpers map
	if err := parseHelpersFlags(helpersFlags, importMap, defaultImport, helperMap); err != nil {
		return nil, err
	}

	// Step 4: Parse legacy -helper flags (highest precedence)
	if err := parseLegacyHelpers(legacyHelperFlags, helperMap); err != nil {
		return nil, err
	}

	return helperMap, nil
}

// parseImports parses -import flags and returns:
// - importMap: map of alias -> import path (only aliased imports)
// - defaultImport: the import path without alias (empty if none)
func parseImports(importFlags importFlag) (map[string]string, string, error) {
	importMap := make(map[string]string)
	var defaultImport string

	for _, raw := range importFlags {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		var importPath, alias string
		if strings.Contains(raw, ":") {
			parts := strings.SplitN(raw, ":", 2)
			alias = strings.TrimSpace(parts[0])
			importPath = strings.TrimSpace(parts[1])
			if alias == "" || importPath == "" {
				return nil, "", fmt.Errorf("invalid import flag %q", raw)
			}
			if _, exists := importMap[alias]; exists {
				return nil, "", fmt.Errorf("duplicate import alias %q", alias)
			}
			importMap[alias] = importPath
		} else {
			importPath = raw
			if defaultImport != "" {
				return nil, "", fmt.Errorf("multiple default imports specified (only one import without alias allowed)")
			}
			defaultImport = importPath
		}
	}

	return importMap, defaultImport, nil
}

// parseHelpersFlags parses -helpers flags and adds helpers to the map.
// Format: [alias:]Name or [alias:]name=Ident
// If name= is omitted, template name defaults to strings.ToLower(Ident)
func parseHelpersFlags(helpersFlags helpersFlag, importMap map[string]string, defaultImport string, helpers map[string]compiler.HelperRef) error {
	for _, raw := range helpersFlags {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		// Split by comma to support multiple helpers per flag
		helperList := strings.Split(raw, ",")
		for _, helperSpec := range helperList {
			helperSpec = strings.TrimSpace(helperSpec)
			if helperSpec == "" {
				continue
			}
			var alias, name, ident string
			var importPath string

			// Check if it has alias prefix (alias:...)
			if strings.Contains(helperSpec, ":") {
				parts := strings.SplitN(helperSpec, ":", 2)
				alias = strings.TrimSpace(parts[0])
				helperSpec = strings.TrimSpace(parts[1])
			}

			// Check if it has name=Ident format
			if strings.Contains(helperSpec, "=") {
				parts := strings.SplitN(helperSpec, "=", 2)
				name = strings.TrimSpace(parts[0])
				ident = strings.TrimSpace(parts[1])
			} else {
				// No name=, use the spec as Ident and default name to lowercase
				ident = helperSpec
				name = strings.ToLower(ident)
			}

			if name == "" || ident == "" {
				return fmt.Errorf("invalid helper spec %q", helperSpec)
			}

			// Determine import path
			if alias != "" {
				var ok bool
				importPath, ok = importMap[alias]
				if !ok {
					return fmt.Errorf("unknown import alias %q", alias)
				}
			} else if defaultImport != "" {
				importPath = defaultImport
			} else {
				return fmt.Errorf("helper %q requires an import alias or default import", helperSpec)
			}

			helpers[name] = compiler.HelperRef{
				ImportPath: importPath,
				Ident:      ident,
			}
		}
	}
	return nil
}

// parseLegacyHelpers parses the legacy -helper flags (for backward compatibility)
func parseLegacyHelpers(values []string, helpers map[string]compiler.HelperRef) error {
	for _, raw := range values {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid helper mapping %q", raw)
		}
		name := strings.TrimSpace(parts[0])
		ref := strings.TrimSpace(parts[1])
		if name == "" || ref == "" {
			return fmt.Errorf("invalid helper mapping %q", raw)
		}
		var importPath, ident string
		if strings.Contains(ref, ":") {
			refParts := strings.SplitN(ref, ":", 2)
			importPath = strings.TrimSpace(refParts[0])
			ident = strings.TrimSpace(refParts[1])
			if importPath == "" || ident == "" {
				return fmt.Errorf("invalid helper mapping %q", raw)
			}
		} else {
			ident = ref
		}
		// Legacy flags can override existing helpers
		helpers[name] = compiler.HelperRef{ImportPath: importPath, Ident: ident}
	}
	return nil
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
