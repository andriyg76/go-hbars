package sitegen

import (
	"io"

	"github.com/andriyg76/go-hbars/internal/processor"
	"github.com/andriyg76/go-hbars/pkg/renderer"
	"github.com/andriyg76/hexerr"
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
// NewRendererFromFunctions accepts map[string]func(io.Writer, any) error so that
// generated bootstrap code (rendererFuncs) can be passed without type conversion.
func NewRendererFromFunctions(funcs map[string]func(io.Writer, any) error) renderer.TemplateRenderer {
	r, _ := processor.NewCompiledTemplateRenderer(funcs)
	return r
}

// LoadRendererFromPackage attempts to load render functions from a package using reflection.
// This works when the package has a struct with Render* methods or when you provide
// a map of functions.
//
// For packages with standalone Render* functions, use NewRendererFromFunctions instead.
func LoadRendererFromPackage(templatePackage any) (renderer.TemplateRenderer, error) {
	r, err := processor.NewCompiledTemplateRenderer(templatePackage)
	if err != nil {
		return nil, hexerr.Wrap(err, "failed to create renderer")
	}
	return r, nil
}

// AutoLoadRenderer attempts to automatically discover and load render functions.
// It tries multiple strategies:
// 1. If templatePackage is a map[string]RenderFunc, uses it directly
// 2. If templatePackage is a struct with Render* methods, uses reflection
// 3. Otherwise returns an error
func AutoLoadRenderer(templatePackage any) (renderer.TemplateRenderer, error) {
	if templatePackage == nil {
		return nil, hexerr.New("templatePackage cannot be nil")
	}

	// Check if it's already a map of functions (convert to unnamed type for NewRendererFromFunctions)
	if funcMap, ok := templatePackage.(map[string]RenderFunc); ok {
		m := make(map[string]func(io.Writer, any) error, len(funcMap))
		for k, v := range funcMap {
			m[k] = v
		}
		return NewRendererFromFunctions(m), nil
	}

	// Try reflection-based loading
	return LoadRendererFromPackage(templatePackage)
}
