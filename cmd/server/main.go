package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andriyg76/go-hbars/internal/processor"
	"github.com/andriyg76/go-hbars/internal/server"
)

func main() {
	var (
		rootPath      = flag.String("root", "", "base directory for resolving relative paths (default: current working directory)")
		dataPath      = flag.String("data-path", "data", "path to data files directory")
		sharedPath    = flag.String("shared-path", "shared", "path to shared data directory")
		templatesPath = flag.String("templates-path", ".processor/templates", "path to templates directory")
		staticDir     = flag.String("static-dir", "", "path to static files directory (optional)")
		addr          = flag.String("addr", ":8080", "address to listen on")
		templatePkg   = flag.String("template-pkg", "", "path to compiled template package (e.g., github.com/your/project/templates)")
	)
	flag.Parse()

	// Determine root path
	root := *rootPath
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
	}

	// Create configuration
	config := &processor.Config{
		RootPath:      root,
		DataPath:      *dataPath,
		SharedPath:    *sharedPath,
		TemplatesPath: *templatesPath,
		OutputPath:    "", // Not used for server
	}

	// Load shared data
	sharedData, err := processor.LoadSharedData(filepath.Join(root, *sharedPath))
	if err != nil {
		log.Fatalf("Failed to load shared data: %v", err)
	}

	// Create renderer
	// Note: In a real implementation, you would load the compiled template package
	// For now, we'll create a placeholder that needs to be implemented
	renderer, err := createRenderer(*templatePkg)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}

	// Create processor
	proc := processor.NewProcessor(config, renderer)

	// Create handler
	handler := server.NewHandler(proc, sharedData, filepath.Join(root, *staticDir))

	// Start server
	fmt.Printf("Starting server on %s\n", *addr)
	fmt.Printf("Data path: %s\n", filepath.Join(root, *dataPath))
	fmt.Printf("Templates path: %s\n", filepath.Join(root, *templatesPath))
	if *staticDir != "" {
		fmt.Printf("Static files: %s\n", filepath.Join(root, *staticDir))
	}
	fmt.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// createRenderer creates a template renderer.
// This is a placeholder - in a real implementation, you would load the compiled template package.
func createRenderer(templatePkg string) (processor.TemplateRenderer, error) {
	// For now, return an error indicating that templates need to be loaded
	// In a real implementation, you would use reflection or a registry to load
	// the compiled template package
	return nil, fmt.Errorf("template renderer not implemented - you need to load compiled templates")
}

