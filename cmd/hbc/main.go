package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

func main() {
	in := flag.String("in", "", "input template file or directory (required)")
	out := flag.String("out", "", "output Go file path (default: templates_gen.go)")
	pkg := flag.String("pkg", "", "package name for generated code (default: derived from output path)")
	bootstrap := flag.Bool("bootstrap", false, "generate NewQuickServer and NewQuickProcessor")
	ext := flag.String("ext", ".hbs,.handlebars", "comma-separated template extensions")
	flag.Parse()

	if *in == "" {
		fmt.Fprintln(os.Stderr, "hbc: -in is required")
		os.Exit(1)
	}

	templates, templateFiles, err := loadTemplates(*in, strings.Split(*ext, ","))
	if err != nil {
		fmt.Fprintf(os.Stderr, "hbc: load templates: %v\n", err)
		os.Exit(1)
	}

	if len(templates) == 0 {
		fmt.Fprintln(os.Stderr, "hbc: no templates found in", *in)
		os.Exit(1)
	}

	outPath := *out
	if outPath == "" {
		outPath = "templates_gen.go"
	}
	pkgName := *pkg
	if pkgName == "" {
		pkgName = filepath.Base(filepath.Dir(outPath))
		if pkgName == "." {
			pkgName = "templates"
		}
	}

	helperRegistry := helpers.Registry()
	compilerHelpers := make(map[string]compiler.HelperRef, len(helperRegistry))
	for name, ref := range helperRegistry {
		compilerHelpers[name] = compiler.HelperRef{
			ImportPath: ref.ImportPath,
			Ident:      ref.Ident,
		}
	}

	opts := compiler.Options{
		PackageName:       pkgName,
		GenerateBootstrap: *bootstrap,
		Helpers:           compilerHelpers,
		TemplateFiles:     templateFiles,
	}

	code, err := compiler.CompileTemplates(templates, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hbc: compile: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, code, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "hbc: write %s: %v\n", outPath, err)
		os.Exit(1)
	}
}

func loadTemplates(in string, exts []string) (map[string]string, map[string]string, error) {
	info, err := os.Stat(in)
	if err != nil {
		return nil, nil, err
	}
	if info.IsDir() {
		return loadTemplatesDir(in, in, exts)
	}
	content, err := os.ReadFile(in)
	if err != nil {
		return nil, nil, err
	}
	name := filepath.Base(in)
	for _, e := range exts {
		name = strings.TrimSuffix(name, strings.TrimSpace(e))
	}
	templates := map[string]string{name: string(content)}
	files := map[string]string{name: in}
	return templates, files, nil
}

func loadTemplatesDir(root, dir string, exts []string) (map[string]string, map[string]string, error) {
	templates := make(map[string]string)
	templateFiles := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := info.Name()
		var hasExt bool
		for _, e := range exts {
			e = strings.TrimSpace(e)
			if strings.HasSuffix(base, e) {
				hasExt = true
				break
			}
		}
		if !hasExt {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		for _, e := range exts {
			e = strings.TrimSpace(e)
			name = strings.TrimSuffix(name, e)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		templates[name] = string(content)
		templateFiles[name] = path
		return nil
	})
	return templates, templateFiles, err
}
