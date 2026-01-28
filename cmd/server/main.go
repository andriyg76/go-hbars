package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andriyg76/glog"
	"github.com/andriyg76/go-hbars/internal/processor"
	"github.com/andriyg76/go-hbars/internal/server"
	"github.com/andriyg76/hexerr"
)

type e struct{}

func main() {
	hexerr.SetFilterPrefixes(e{})

	log := glog.Create(glog.INFO)

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
			log.Fatal("Failed to get working directory: %v", err)
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
		log.Fatal("Failed to load shared data: %v", err)
	}

	// Create renderer
	// Note: In a real implementation, you would load the compiled template package
	// For now, we'll create a placeholder that needs to be implemented
	renderer, err := createRenderer(*templatePkg)
	if err != nil {
		log.Fatal("Failed to create renderer: %v", err)
	}

	// Create processor
	proc := processor.NewProcessor(config, renderer)

	// Create handler
	handler := server.NewHandler(proc, sharedData, filepath.Join(root, *staticDir))

	// Start server
	log.Info("Starting server on %s", *addr)
	log.Info("Data path: %s", filepath.Join(root, *dataPath))
	log.Info("Templates path: %s", filepath.Join(root, *templatesPath))
	if *staticDir != "" {
		log.Info("Static files: %s", filepath.Join(root, *staticDir))
	}
	log.Info("Press Ctrl+C to stop")

	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("Server failed: %v", err)
	}
}

// createRenderer creates a template renderer.
// This is a placeholder - in a real implementation, you would load the compiled template package.
func createRenderer(templatePkg string) (processor.TemplateRenderer, error) {
	// For now, return an error indicating that templates need to be loaded
	// In a real implementation, you would use reflection or a registry to load
	// the compiled template package
	return nil, hexerr.New("template renderer not implemented - you need to load compiled templates")
}
