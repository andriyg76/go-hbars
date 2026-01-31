package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/andriyg76/glog"
	"github.com/andriyg76/go-hbars/internal/processor"
	"github.com/andriyg76/go-hbars/pkg/renderer"
	"github.com/andriyg76/hexerr"
)

func main() {
	hexerr.SetFilterPrefixes("github.com/andriyg76/go-hbars")

	log := glog.Create(glog.INFO)

	var (
		rootPath      = flag.String("root", "", "base directory for resolving relative paths (default: current working directory)")
		dataPath      = flag.String("data-path", "data", "path to data files directory")
		sharedPath    = flag.String("shared-path", "shared", "path to shared data directory")
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
			log.Fatal("Failed to get working directory: %v", err)
		}
	}

	// Create configuration
	config := &processor.Config{
		RootPath:   root,
		DataPath:   *dataPath,
		SharedPath: *sharedPath,
		OutputPath: *outputPath,
	}

	// Create renderer
	renderer, err := createRenderer(*templatePkg)
	if err != nil {
		log.Fatal("Failed to create renderer: %v", err)
	}

	// Create processor
	proc := processor.NewProcessor(config, renderer)

	// Process all files
	log.Info("Processing files from %s", filepath.Join(root, *dataPath))
	log.Info("Output directory: %s", filepath.Join(root, *outputPath))

	if err := proc.Process(); err != nil {
		log.Fatal("Failed to process files: %v", err)
	}

	log.Info("Done!")
}

// createRenderer creates a template renderer.
// This is a placeholder - in a real implementation, you would load the compiled template package.
func createRenderer(templatePkg string) (renderer.TemplateRenderer, error) {
	// For now, return an error indicating that templates need to be loaded
	return nil, hexerr.New("template renderer not implemented - you need to load compiled templates")
}
