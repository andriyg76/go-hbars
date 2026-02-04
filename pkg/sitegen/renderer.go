package sitegen

import (
	"github.com/andriyg76/go-hbars/pkg/renderer"
)

// LoadRenderer creates a renderer from a compiled template package.
// templatePackage must be map[string]RenderFunc with template functions (nil returns error).
// For standalone Render* functions, use NewRendererFromFunctions instead.
//
// Example:
//
//	renderer, err := sitegen.LoadRenderer(map[string]sitegen.RenderFunc{
//	    "main": templates.RenderMain,
//	})
//	// or: renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{"main": templates.RenderMain})
func LoadRenderer(templatePackage any) (renderer.TemplateRenderer, error) {
	return AutoLoadRenderer(templatePackage)
}

