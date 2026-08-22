package sitebuild

import (
	"fmt"
	"strings"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

// CompileOptions configures Go source generation for a template directory.
type CompileOptions struct {
	TemplatesPath     string
	PackageName       string
	GenerateBootstrap bool
}

// Compile loads a template directory and returns its generated Go source.
func Compile(opts CompileOptions) ([]byte, error) {
	if strings.TrimSpace(opts.TemplatesPath) == "" {
		return nil, fmt.Errorf("templates path is required")
	}
	if strings.TrimSpace(opts.PackageName) == "" {
		opts.PackageName = "templates"
	}

	templates, templateFiles, err := compiler.LoadTemplates(opts.TemplatesPath)
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}
	if len(templates) == 0 {
		return nil, fmt.Errorf("no .hbs templates found in %s", opts.TemplatesPath)
	}

	registry := helpers.Registry()
	compilerHelpers := make(map[string]compiler.HelperRef, len(registry))
	for name, ref := range registry {
		compilerHelpers[name] = compiler.HelperRef{ImportPath: ref.ImportPath, Ident: ref.Ident}
	}
	generated, err := compiler.CompileTemplates(templates, compiler.Options{
		PackageName:       opts.PackageName,
		Helpers:           compilerHelpers,
		GenerateBootstrap: opts.GenerateBootstrap,
		TemplateFiles:     templateFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("compile templates: %w", err)
	}
	return generated, nil
}
