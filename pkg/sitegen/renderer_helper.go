package sitegen

import (
	"fmt"
	"io"

	"github.com/andriyg76/go-hbars/internal/processor"
)

// RenderFunc is a function that renders a template.
type RenderFunc func(io.Writer, any) error

// NewRendererFromFunctions creates a renderer from a map of template names to render functions.
// This is useful when you have direct access to the compiled template functions.
//
// Example:
//
//	import "github.com/your/project/templates"
//	renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
//	    "main":   templates.RenderMain,
//	    "header": templates.RenderHeader,
//	    "footer": templates.RenderFooter,
//	})
func NewRendererFromFunctions(funcs map[string]RenderFunc) processor.TemplateRenderer {
	renderer, _ := processor.NewCompiledTemplateRenderer(funcs)
	return renderer
}

// LoadRendererFromPackage attempts to load render functions from a package using reflection.
// This works when the package has a struct with Render* methods or when you provide
// a map of functions.
//
// For packages with standalone Render* functions, use NewRendererFromFunctions instead.
func LoadRendererFromPackage(templatePackage any) (processor.TemplateRenderer, error) {
	renderer, err := processor.NewCompiledTemplateRenderer(templatePackage)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}
	return renderer, nil
}

// AutoLoadRenderer attempts to automatically discover and load render functions.
// It tries multiple strategies:
// 1. If templatePackage is a map[string]RenderFunc, uses it directly
// 2. If templatePackage is a struct with Render* methods, uses reflection
// 3. Otherwise returns an error
func AutoLoadRenderer(templatePackage any) (processor.TemplateRenderer, error) {
	if templatePackage == nil {
		return nil, fmt.Errorf("templatePackage cannot be nil")
	}

	// Check if it's already a map of functions
	if funcMap, ok := templatePackage.(map[string]RenderFunc); ok {
		return NewRendererFromFunctions(funcMap), nil
	}

	// Try reflection-based loading
	return LoadRendererFromPackage(templatePackage)
}

