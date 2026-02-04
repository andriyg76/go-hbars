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
	return processor.NewCompiledTemplateRenderer(funcs)
}

// LoadRendererFromPackage creates a renderer from a map of render functions.
// For standalone Render* functions, use NewRendererFromFunctions instead.
func LoadRendererFromPackage(funcs map[string]func(io.Writer, any) error) renderer.TemplateRenderer {
	return processor.NewCompiledTemplateRenderer(funcs)
}

// AutoLoadRenderer creates a renderer from templatePackage.
// templatePackage must be map[string]RenderFunc (or map[string]func(io.Writer, any) error).
// Struct with Render* methods is no longer supported; use NewRendererFromFunctions with a map.
func AutoLoadRenderer(templatePackage any) (renderer.TemplateRenderer, error) {
	if templatePackage == nil {
		return nil, hexerr.New("templatePackage cannot be nil")
	}

	if funcMap, ok := templatePackage.(map[string]RenderFunc); ok {
		m := make(map[string]func(io.Writer, any) error, len(funcMap))
		for k, v := range funcMap {
			m[k] = v
		}
		return NewRendererFromFunctions(m), nil
	}

	return nil, hexerr.New("templatePackage must be map[string]RenderFunc (or map[string]func(io.Writer, any) error); struct with Render* methods is no longer supported")
}
