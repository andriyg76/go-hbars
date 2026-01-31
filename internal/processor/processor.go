package processor

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/andriyg76/go-hbars/pkg/renderer"
	"github.com/andriyg76/hexerr"
)

// Config holds processor configuration.
type Config struct {
	RootPath   string
	DataPath   string
	SharedPath string
	OutputPath string
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		DataPath:   "data",
		SharedPath: "shared",
		OutputPath: "pages",
	}
}

// Processor processes data files and renders them using templates.
type Processor struct {
	config   *Config
	renderer renderer.TemplateRenderer
}

// NewProcessor creates a new processor with the given configuration and renderer.
func NewProcessor(config *Config, r renderer.TemplateRenderer) *Processor {
	return &Processor{
		config:   config,
		renderer: r,
	}
}

// Config returns the processor configuration.
func (p *Processor) Config() *Config {
	return p.config
}

// Process processes all data files and generates output files.
func (p *Processor) Process() error {
	// Load shared data
	sharedPath := p.resolvePath(p.config.SharedPath)
	sharedData, err := LoadSharedData(sharedPath)
	if err != nil {
		return hexerr.Wrapf(err, "failed to load shared data")
	}

	// Process data files
	dataPath := p.resolvePath(p.config.DataPath)
	if err := p.processDirectory(dataPath, sharedData, ""); err != nil {
		return hexerr.Wrapf(err, "failed to process data files")
	}

	return nil
}

// ProcessFile processes a single data file.
func (p *Processor) ProcessFile(dataFilePath string, sharedData map[string]any) (string, []byte, error) {
	// Load page data
	pageData, err := LoadDataFile(dataFilePath)
	if err != nil {
		return "", nil, err
	}

	// Extract page config
	pageConfig, err := ExtractPageConfig(pageData)
	if err != nil {
		return "", nil, err
	}
	if pageConfig == nil {
		return "", nil, nil // File should be ignored
	}

	// Remove _page from data
	RemovePageConfig(pageData)

	// Merge shared data
	MergeSharedData(pageData, sharedData)

	// Render template
	var buf strings.Builder
	if err := p.renderer.Render(pageConfig.Template, &buf, pageData); err != nil {
		return "", nil, hexerr.Wrapf(err, "failed to render template %q", pageConfig.Template)
	}

	// Determine output path
	outputPath := p.determineOutputPath(dataFilePath, pageConfig)

	return outputPath, []byte(buf.String()), nil
}

func (p *Processor) processDirectory(dirPath string, sharedData map[string]any, relPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return hexerr.Wrapf(err, "failed to read directory %q", dirPath)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		currentRelPath := filepath.Join(relPath, entry.Name())

		if entry.IsDir() {
			// Recursively process subdirectories
			if err := p.processDirectory(fullPath, sharedData, currentRelPath); err != nil {
				return err
			}
			continue
		}

		// Process data files
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".json5" && ext != ".yaml" && ext != ".yml" && ext != ".toml" {
			continue
		}

		outputPath, content, err := p.ProcessFile(fullPath, sharedData)
		if err != nil {
			return hexerr.Wrapf(err, "failed to process file %q", fullPath)
		}
		if outputPath == "" {
			continue // File should be ignored
		}

		// Write output file
		if err := p.writeOutputFile(outputPath, content); err != nil {
			return hexerr.Wrapf(err, "failed to write output file %q", outputPath)
		}
	}

	return nil
}

func (p *Processor) determineOutputPath(dataFilePath string, pageConfig *PageConfig) string {
	if pageConfig.Output != "" {
		outputPath := pageConfig.Output
		// Remove leading slash if present
		if strings.HasPrefix(outputPath, "/") {
			outputPath = outputPath[1:]
		}
		return p.resolvePath(p.config.OutputPath, outputPath)
	}

	// Default: use input file name with .html extension
	baseName := filepath.Base(dataFilePath)
	ext := filepath.Ext(baseName)
	baseName = strings.TrimSuffix(baseName, ext)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName)) // Handle .json5

	// Get relative path from data directory
	dataPath := p.resolvePath(p.config.DataPath)
	relPath, err := filepath.Rel(dataPath, dataFilePath)
	if err == nil && relPath != "." {
		dir := filepath.Dir(relPath)
		if dir != "." {
			return p.resolvePath(p.config.OutputPath, dir, baseName+".html")
		}
	}

	return p.resolvePath(p.config.OutputPath, baseName+".html")
}

func (p *Processor) writeOutputFile(outputPath string, content []byte) error {
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return hexerr.Wrapf(err, "failed to create output directory %q", dir)
		}
	}

	if err := os.WriteFile(outputPath, content, 0o644); err != nil {
		return hexerr.Wrapf(err, "failed to write file %q", outputPath)
	}

	return nil
}

func (p *Processor) resolvePath(parts ...string) string {
	path := filepath.Join(parts...)
	if filepath.IsAbs(path) {
		return path
	}
	if p.config.RootPath != "" {
		return filepath.Join(p.config.RootPath, path)
	}
	return path
}
