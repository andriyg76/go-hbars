package sitegen

import (
	"os"
	"path/filepath"

	"github.com/andriyg76/go-hbars/internal/processor"
	"github.com/andriyg76/go-hbars/pkg/renderer"
	"github.com/andriyg76/hexerr"
)

// Processor processes data files and generates static HTML files.
type Processor struct {
	config   *Config
	proc     *processor.Processor
	renderer renderer.TemplateRenderer
}

// NewProcessor creates a new processor with the given configuration and renderer.
func NewProcessor(config *Config, r renderer.TemplateRenderer) (*Processor, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Determine root path
	root := config.RootPath
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, hexerr.Wrap(err, "failed to get working directory")
		}
	}

	procConfig := &processor.Config{
		RootPath:   root,
		DataPath:   config.DataPath,
		SharedPath: config.SharedPath,
		OutputPath: config.OutputPath,
	}

	proc := processor.NewProcessor(procConfig, r)

	return &Processor{
		config:   config,
		proc:     proc,
		renderer: r,
	}, nil
}

// Process processes all data files and generates output files.
func (p *Processor) Process() error {
	return p.proc.Process()
}

// ProcessFile processes a single data file and returns the output path and content.
func (p *Processor) ProcessFile(dataFilePath string) (string, []byte, error) {
	// Load shared data
	sharedPath := filepath.Join(p.config.RootPath, p.config.SharedPath)
	sharedData, err := processor.LoadSharedData(sharedPath)
	if err != nil {
		return "", nil, hexerr.Wrap(err, "failed to load shared data")
	}

	return p.proc.ProcessFile(dataFilePath, sharedData)
}

// Config returns the processor configuration.
func (p *Processor) Config() *Config {
	return p.config
}
