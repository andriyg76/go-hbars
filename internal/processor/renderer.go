package processor

import (
	"fmt"
	"io"
	"strings"

	"github.com/andriyg76/hexerr"
)

// RenderFunc is the type of a function that renders a template: (w, data) -> error.
// Using a type alias keeps compatibility with func(io.Writer, any) error.
type RenderFunc = func(io.Writer, any) error

// CompiledTemplateRenderer renders templates using compiled template functions.
// It is created from a map of names to RenderFunc (e.g. bootstrap rendererFuncs)
// or nil with manual RegisterRenderFunc; it calls functions directly without reflection.
type CompiledTemplateRenderer struct {
	funcs map[string]RenderFunc
}

// NewCompiledTemplateRenderer creates a new renderer from a map of template names to render functions.
// funcs may be nil (empty renderer; use RegisterRenderFunc to add) or a non-nil map.
func NewCompiledTemplateRenderer(funcs map[string]RenderFunc) *CompiledTemplateRenderer {
	r := &CompiledTemplateRenderer{}
	if len(funcs) == 0 {
		return r
	}
	r.funcs = make(map[string]RenderFunc, len(funcs))
	for name, fn := range funcs {
		r.funcs[normalizeTemplateName(name)] = fn
	}
	return r
}

// Render renders a template by name.
func (r *CompiledTemplateRenderer) Render(templateName string, w io.Writer, data any) error {
	actualName, ok := r.findTemplateName(templateName)
	if !ok {
		return hexerr.New(fmt.Sprintf("template %q not found in compiled templates (available: %v)", templateName, r.getAvailableTemplates()))
	}

	fn := r.funcs[actualName]
	if fn == nil {
		return hexerr.New(fmt.Sprintf("template %q not found in compiled templates", templateName))
	}
	if err := fn(w, data); err != nil {
		return hexerr.Wrapf(err, "failed to render template %q", templateName)
	}
	return nil
}

// normalizeTemplateName converts template paths to function names.
// Examples:
//   - "blog/post" -> tries "blogpost", "blog/post", "post"
//   - "main" -> "main"
func normalizeTemplateName(name string) string {
	name = strings.Trim(name, "/\\")
	return name
}

// hasTemplate returns true if name exists in funcs.
func (r *CompiledTemplateRenderer) hasTemplate(name string) bool {
	return r.funcs != nil && r.funcs[name] != nil
}

// findTemplateName tries to find a matching template name in the render funcs map.
func (r *CompiledTemplateRenderer) findTemplateName(templateName string) (string, bool) {
	normalized := normalizeTemplateName(templateName)

	if r.hasTemplate(normalized) {
		return normalized, true
	}

	noPath := strings.ReplaceAll(normalized, "/", "")
	noPath = strings.ReplaceAll(noPath, "\\", "")
	if r.hasTemplate(noPath) {
		return noPath, true
	}

	parts := strings.Split(normalized, "/")
	if len(parts) > 1 {
		baseName := parts[len(parts)-1]
		if r.hasTemplate(baseName) {
			return baseName, true
		}
		camelCase := parts[0]
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > 0 {
				camelCase += strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
		if r.hasTemplate(camelCase) {
			return camelCase, true
		}
	}

	return "", false
}

// RegisterRenderFunc registers a render function manually (direct call, no reflection).
func (r *CompiledTemplateRenderer) RegisterRenderFunc(templateName string, renderFunc RenderFunc) {
	name := normalizeTemplateName(templateName)
	if r.funcs == nil {
		r.funcs = make(map[string]RenderFunc)
	}
	r.funcs[name] = renderFunc
}

// getAvailableTemplates returns a list of available template names.
func (r *CompiledTemplateRenderer) getAvailableTemplates() []string {
	if r.funcs == nil {
		return nil
	}
	names := make([]string, 0, len(r.funcs))
	for name := range r.funcs {
		names = append(names, name)
	}
	return names
}
