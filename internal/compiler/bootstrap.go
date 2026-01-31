package compiler

// generateBootstrapCode generates helper functions for quick server/processor setup.
// It writes rendererFuncs, NewRenderer, NewQuickProcessor, and NewQuickServer.
// partialParamTypes: when a partial uses another template's context (e.g. footer uses MainContext),
// the bootstrap uses that context type and its FromMap for the wrapper.
// useLayoutBlocks: when true, templates use {{#block}}/{{#partial}} layout and rendererFuncs call RenderXxxWithBlocks with runtime.NewBlocks().
func generateBootstrapCode(w *codeWriter, templateNames []string, funcNames map[string]string, partialParamTypes map[string]string, useLayoutBlocks bool) {
	// Generate renderer map with wrappers that accept any and convert to context type
	w.line("")
	w.line("// rendererFuncs maps template names to render functions.")
	w.line("var rendererFuncs = map[string]func(io.Writer, any) error{")
	w.indentInc()
	for _, name := range templateNames {
		goName := funcNames[name]
		rootContext := partialParamTypes[name]
		if rootContext == "" {
			rootContext = goName + "Context"
		}
		fromMap := rootContext + "FromMap"
		w.line("%q: func(w io.Writer, data any) error {", name)
		w.indentInc()
		if useLayoutBlocks {
			w.line("blocks := runtime.NewBlocks()")
			w.line("if c, ok := data.(%s); ok { return Render%sWithBlocks(w, c, blocks) }", rootContext, goName)
			w.line("if m, ok := data.(map[string]any); ok { return Render%sWithBlocks(w, %s(m), blocks) }", goName, fromMap)
		} else {
			w.line("if c, ok := data.(%s); ok { return Render%s(w, c) }", rootContext, goName)
			w.line("if m, ok := data.(map[string]any); ok { return Render%s(w, %s(m)) }", goName, fromMap)
		}
		w.line("return fmt.Errorf(%q, data)", name+": expected "+rootContext+" or map[string]any, got %T")
		w.indentDec()
		w.line("},")
	}
	w.indentDec()
	w.line("}")

	// Generate NewRenderer function
	w.line("")
	w.line("// NewRenderer returns a ready-to-use template renderer.")
	w.line("// This renderer can be used with sitegen.NewProcessor or sitegen.NewServer.")
	w.line("func NewRenderer() renderer.TemplateRenderer {")
	w.indentInc()
	w.line("return sitegen.NewRendererFromFunctions(rendererFuncs)")
	w.indentDec()
	w.line("}")

	// Generate quick processor function
	w.line("")
	w.line("// NewQuickProcessor creates a processor with default configuration.")
	w.line("// Use this for quick static site generation.")
	w.line("func NewQuickProcessor() (*sitegen.Processor, error) {")
	w.indentInc()
	w.line("config := sitegen.DefaultConfig()")
	w.line("renderer := NewRenderer()")
	w.line("return sitegen.NewProcessor(config, renderer)")
	w.indentDec()
	w.line("}")

	// Generate quick server function
	w.line("")
	w.line("// NewQuickServer creates a server with default configuration.")
	w.line("// Use this for quick development server setup.")
	w.line("func NewQuickServer() (*sitegen.Server, error) {")
	w.indentInc()
	w.line("config := sitegen.DefaultConfig()")
	w.line("renderer := NewRenderer()")
	w.line("return sitegen.NewServer(config, renderer)")
	w.indentDec()
	w.line("}")
}
