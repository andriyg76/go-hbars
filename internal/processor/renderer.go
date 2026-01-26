package processor

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// CompiledTemplateRenderer renders templates using compiled template functions.
// It uses reflection to call Render* functions from compiled template packages.
type CompiledTemplateRenderer struct {
	templatePackage reflect.Value
	renderFuncs     map[string]reflect.Value
}

// NewCompiledTemplateRenderer creates a new renderer that uses compiled templates.
// templatePackage can be:
//   - An instance of a struct with Render* methods (e.g., templates.Templates{})
//   - A package-level function registry (map[string]func(io.Writer, any) error)
//   - nil, in which case you must register functions manually
func NewCompiledTemplateRenderer(templatePackage any) (*CompiledTemplateRenderer, error) {
	renderer := &CompiledTemplateRenderer{
		renderFuncs: make(map[string]reflect.Value),
	}

	if templatePackage == nil {
		return renderer, nil
	}

	// Try to extract functions from the package
	switch v := templatePackage.(type) {
	case map[string]func(io.Writer, any) error:
		// Direct function map
		for name, fn := range v {
			renderer.renderFuncs[normalizeTemplateName(name)] = reflect.ValueOf(fn)
		}
	default:
		// Try reflection on struct/interface
		pkgValue := reflect.ValueOf(templatePackage)
		if pkgValue.Kind() == reflect.Ptr {
			if pkgValue.IsNil() {
				return renderer, nil
			}
			pkgValue = pkgValue.Elem()
		}

		renderer.templatePackage = pkgValue

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
					renderer.renderFuncs[templateName] = pkgValue.Method(i)
				}
			}
		}
	}

	return renderer, nil
}

// Render renders a template by name.
func (r *CompiledTemplateRenderer) Render(templateName string, w io.Writer, data any) error {
	// Find matching template name
	actualName, ok := r.findTemplateName(templateName)
	if !ok {
		return fmt.Errorf("template %q not found in compiled templates (available: %v)", templateName, r.getAvailableTemplates())
	}

	renderFunc, ok := r.renderFuncs[actualName]
	if !ok {
		return fmt.Errorf("template %q not found in compiled templates", templateName)
	}

	// Call the render function: RenderTemplateName(w io.Writer, data any) error
	args := []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(data),
	}

	results := renderFunc.Call(args)
	if len(results) > 0 {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			return fmt.Errorf("failed to render template %q: %w", templateName, err)
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

// findTemplateName tries to find a matching template name in the render funcs map.
// It tries multiple variations:
// 1. Exact match (with path separators removed)
// 2. Base name only (last component)
// 3. All path separators removed
func (r *CompiledTemplateRenderer) findTemplateName(templateName string) (string, bool) {
	normalized := normalizeTemplateName(templateName)
	
	// Try exact normalized match
	if _, ok := r.renderFuncs[normalized]; ok {
		return normalized, true
	}
	
	// Try with path separators removed
	noPath := strings.ReplaceAll(normalized, "/", "")
	noPath = strings.ReplaceAll(noPath, "\\", "")
	if _, ok := r.renderFuncs[noPath]; ok {
		return noPath, true
	}
	
	// Try base name only (last component)
	parts := strings.Split(normalized, "/")
	if len(parts) > 1 {
		baseName := parts[len(parts)-1]
		if _, ok := r.renderFuncs[baseName]; ok {
			return baseName, true
		}
	}
	
	// Try camelCase version (e.g., "blog/post" -> "blogPost")
	if len(parts) > 1 {
		camelCase := parts[0]
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > 0 {
				camelCase += strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
		if _, ok := r.renderFuncs[camelCase]; ok {
			return camelCase, true
		}
	}
	
	return "", false
}

// RegisterRenderFunc registers a render function manually.
// This is useful when you want to register functions directly without reflection.
func (r *CompiledTemplateRenderer) RegisterRenderFunc(templateName string, renderFunc func(io.Writer, any) error) {
	r.renderFuncs[normalizeTemplateName(templateName)] = reflect.ValueOf(renderFunc)
}

// getAvailableTemplates returns a list of available template names.
func (r *CompiledTemplateRenderer) getAvailableTemplates() []string {
	names := make([]string, 0, len(r.renderFuncs))
	for name := range r.renderFuncs {
		names = append(names, name)
	}
	return names
}

