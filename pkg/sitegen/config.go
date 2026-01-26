package sitegen

// Config holds configuration for site generation.
type Config struct {
	// RootPath is the base directory for resolving relative paths.
	// If empty, uses current working directory.
	RootPath string

	// DataPath is the path to data files directory (default: "data").
	DataPath string

	// SharedPath is the path to shared data directory (default: "shared").
	SharedPath string

	// TemplatesPath is the path to templates directory (default: ".processor/templates").
	TemplatesPath string

	// OutputPath is the path to output directory for static generation (default: "pages").
	OutputPath string

	// StaticDir is the path to static files directory for server (optional).
	StaticDir string

	// Addr is the address to listen on for server (default: ":8080").
	Addr string
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		DataPath:      "data",
		SharedPath:    "shared",
		TemplatesPath: ".processor/templates",
		OutputPath:    "pages",
		Addr:          ":8080",
	}
}

