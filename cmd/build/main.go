package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/andriyg76/go-hbars/internal/processor"
)

func main() {
	var (
		rootPath      = flag.String("root", "", "base directory for resolving relative paths (default: current working directory)")
		dataPath      = flag.String("data-path", "data", "path to data files directory")
		sharedPath    = flag.String("shared-path", "shared", "path to shared data directory")
		templatesPath = flag.String("templates-path", ".processor/templates", "path to templates directory")
		outputPath    = flag.String("output-path", "pages", "path to output directory")
		templatePkg   = flag.String("template-pkg", "", "path to compiled template package")
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
		OutputPath:    *outputPath,
	}

	// Create renderer
	renderer, err := createRenderer(*templatePkg)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}

	// Create processor
	proc := processor.NewProcessor(config, renderer)

	// Process all files
	fmt.Printf("Processing files from %s\n", filepath.Join(root, *dataPath))
	fmt.Printf("Output directory: %s\n", filepath.Join(root, *outputPath))

	if err := proc.Process(); err != nil {
		log.Fatalf("Failed to process files: %v", err)
	}

	fmt.Println("Done!")
}

// createRenderer creates a template renderer.
// This is a placeholder - in a real implementation, you would load the compiled template package.
func createRenderer(templatePkg string) (processor.TemplateRenderer, error) {
	// For now, return an error indicating that templates need to be loaded
	return nil, fmt.Errorf("template renderer not implemented - you need to load compiled templates")
}

