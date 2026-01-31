package processor

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/andriyg76/hexerr"
)

// CompiledTemplateRenderer renders templates using compiled template functions.
// When created from a map[string]func(io.Writer, any) error (e.g. bootstrap rendererFuncs),
// it calls functions directly without reflection. When created from a struct with Render*
// methods, it uses reflection for discovery and for calling.
type CompiledTemplateRenderer struct {
	templatePackage reflect.Value
	// funcs is set when using a map or RegisterRenderFunc; used for direct calls (no reflection).
	funcs map[string]func(io.Writer, any) error
	// renderFuncs is set when using struct reflection; calls use reflect.Value.Call.
	renderFuncs map[string]reflect.Value
}

// NewCompiledTemplateRenderer creates a new renderer that uses compiled templates.
// templatePackage can be:
//   - An instance of a struct with Render* methods (e.g., templates.Templates{})
//   - A package-level function registry (map[string]func(io.Writer, any) error)
//   - nil, in which case you must register functions manually
func NewCompiledTemplateRenderer(templatePackage any) (*CompiledTemplateRenderer, error) {
	r := &CompiledTemplateRenderer{
		renderFuncs: make(map[string]reflect.Value),
	}

	if templatePackage == nil {
		return r, nil
	}

	// Try to extract functions from the package
	switch v := templatePackage.(type) {
	case map[string]func(io.Writer, any) error:
		// Direct function map (e.g. bootstrap rendererFuncs): no reflection at call time
		r.funcs = make(map[string]func(io.Writer, any) error, len(v))
		for name, fn := range v {
			r.funcs[normalizeTemplateName(name)] = fn
		}
	default:
		// Try reflection on struct/interface
		pkgValue := reflect.ValueOf(templatePackage)
		if pkgValue.Kind() == reflect.Ptr {
			if pkgValue.IsNil() {
				return r, nil
			}
			pkgValue = pkgValue.Elem()
		}

		r.templatePackage = pkgValue

		// Find all Render* functions in the package
		pkgType := pkgValue.Type()
		for i := 0; i < pkgType.NumMethod(); i++ {
			method := pkgType.Method(i)
			if strings.HasPrefix(method.Name, "Render") {
				// Extract template name from method name (e.g., "RenderMain" -> "main")
				templateName := strings.TrimPrefix(method.Name, "Render")
				if len(templateName) > 0 {
					// Convert to lowercase for consistency
					templateName = strings.ToLower(templateName[:1]) + templateName[1:]
					r.renderFuncs[templateName] = pkgValue.Method(i)
				}
			}
		}
	}

	return r, nil
}

// Render renders a template by name.
func (r *CompiledTemplateRenderer) Render(templateName string, w io.Writer, data any) error {
	// Find matching template name
	actualName, ok := r.findTemplateName(templateName)
	if !ok {
		return hexerr.New(fmt.Sprintf("template %q not found in compiled templates (available: %v)", templateName, r.getAvailableTemplates()))
	}

	if fn := r.funcs[actualName]; fn != nil {
		if err := fn(w, data); err != nil {
			return hexerr.Wrapf(err, "failed to render template %q", templateName)
		}
		return nil
	}

	renderFunc, ok := r.renderFuncs[actualName]
	if !ok {
		return hexerr.New(fmt.Sprintf("template %q not found in compiled templates", templateName))
	}

	// Reflection path: struct methods
	args := []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(data),
	}
	results := renderFunc.Call(args)
	if len(results) > 0 {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			return hexerr.Wrapf(err, "failed to render template %q", templateName)
		}
	}
	return nil
}

// normalizeTemplateName converts template paths to function names.
// Examples:
//   - "blog/post" -> tries "blogpost", "blog/post", "post"
//   - "main" -> "main"
func normalizeTemplateName(name string) string {
	// Remove leading/trailing slashes
	name = strings.Trim(name, "/\\")
	return name
}

// hasTemplate returns true if name exists in funcs or renderFuncs.
func (r *CompiledTemplateRenderer) hasTemplate(name string) bool {
	if r.funcs != nil && r.funcs[name] != nil {
		return true
	}
	_, ok := r.renderFuncs[name]
	return ok
}

// findTemplateName tries to find a matching template name in the render funcs map.
// It tries multiple variations:
// 1. Exact match (with path separators removed)
// 2. Base name only (last component)
// 3. All path separators removed
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
func (r *CompiledTemplateRenderer) RegisterRenderFunc(templateName string, renderFunc func(io.Writer, any) error) {
	name := normalizeTemplateName(templateName)
	if r.funcs == nil {
		r.funcs = make(map[string]func(io.Writer, any) error)
	}
	r.funcs[name] = renderFunc
}

// getAvailableTemplates returns a list of available template names.
func (r *CompiledTemplateRenderer) getAvailableTemplates() []string {
	seen := make(map[string]struct{})
	for name := range r.funcs {
		seen[name] = struct{}{}
	}
	for name := range r.renderFuncs {
		seen[name] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	return names
}
