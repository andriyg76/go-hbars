package sitegen

import (
	"github.com/andriyg76/go-hbars/internal/processor"
)

// LoadRenderer creates a renderer from a compiled template package.
// templatePackage can be:
//   - A map[string]RenderFunc with template functions
//   - A struct instance with Render* methods
//   - nil (returns error)
//
// For packages with standalone Render* functions, use NewRendererFromFunctions instead.
//
// Example with struct:
//
//	type Templates struct{}
//	func (Templates) RenderMain(w io.Writer, data any) error { ... }
//
//	renderer, err := sitegen.LoadRenderer(Templates{})
//
// Example with function map:
//
//	renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
//	    "main": templates.RenderMain,
//	})
func LoadRenderer(templatePackage any) (processor.TemplateRenderer, error) {
	return AutoLoadRenderer(templatePackage)
}

