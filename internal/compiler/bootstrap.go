package compiler

// generateBootstrapCode generates helper functions for quick server/processor setup.
// It writes rendererFuncs, NewRenderer, NewQuickProcessor, and NewQuickServer.
func generateBootstrapCode(w *codeWriter, templateNames []string, funcNames map[string]string) {
	// Generate renderer map
	w.line("")
	w.line("// rendererFuncs maps template names to render functions.")
	w.line("var rendererFuncs = map[string]func(io.Writer, any) error{")
	w.indentInc()
	for _, name := range templateNames {
		goName := funcNames[name]
		w.line("%q: Render%s,", name, goName)
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
