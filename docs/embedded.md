# Embedded Processor and Web Server

go-hbars provides APIs for embedding site generation and web server functionality into your Go applications.

## Quick Start with Bootstrap Code

If you used `-bootstrap` flag when generating templates, you can use the quick functions:

### Quick Processor

```go
import "github.com/your/project/templates"

// Quick processor with default config
proc, err := templates.NewQuickProcessor()
if err != nil {
    log.Fatal(err)
}

// Customize config if needed
proc.Config().DataPath = "content"
proc.Config().OutputPath = "build"

if err := proc.Process(); err != nil {
    log.Fatal(err)
}
```

### Quick Server

```go
import "github.com/your/project/templates"

// Quick server with default config
srv, err := templates.NewQuickServer()
if err != nil {
    log.Fatal(err)
}

// Customize config if needed
srv.Config().DataPath = "content"
srv.Config().Addr = ":3000"

log.Fatal(srv.Start())
```

## Using the API Directly

### Static Site Generation

```go
import (
    "github.com/andriyg76/go-hbars/pkg/sitegen"
    "github.com/your/project/templates"
)

config := sitegen.DefaultConfig()
config.DataPath = "data"
config.OutputPath = "pages"

// Create renderer from compiled template functions
renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
    "main":   templates.RenderMain,
    "header": templates.RenderHeader,
    "footer": templates.RenderFooter,
})

proc, err := sitegen.NewProcessor(config, renderer)
if err != nil {
    log.Fatal(err)
}

if err := proc.Process(); err != nil {
    log.Fatal(err)
}
```

### Semi-Static Web Server

```go
import (
    "github.com/andriyg76/go-hbars/pkg/sitegen"
    "github.com/your/project/templates"
)

config := sitegen.DefaultConfig()
config.DataPath = "data"
config.Addr = ":8080"

// Create renderer from compiled template functions
renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
    "main":   templates.RenderMain,
    "header": templates.RenderHeader,
    "footer": templates.RenderFooter,
})

srv, err := sitegen.NewServer(config, renderer)
if err != nil {
    log.Fatal(err)
}

log.Fatal(srv.Start())
```

## API Reference

### Configuration

```go
type Config struct {
    RootPath      string // Base directory for resolving relative paths
    DataPath      string // Path to data files directory (default: "data")
    SharedPath    string // Path to shared data directory (default: "shared")
    TemplatesPath string // Path to templates directory (default: ".processor/templates")
    OutputPath    string // Path to output directory for static generation (default: "pages")
    StaticDir     string // Path to static files directory for server (optional)
    Addr          string // Address to listen on for server (default: ":8080")
}
```

### Processor

```go
// NewProcessor creates a new processor with the given configuration and renderer
proc, err := sitegen.NewProcessor(config, renderer)

// Process processes all data files and generates output files
err := proc.Process()

// ProcessFile processes a single data file and returns the output path and content
outputPath, content, err := proc.ProcessFile(dataFilePath)

// Config returns the processor configuration
config := proc.Config()
```

### Server

```go
// NewServer creates a new server with the given configuration and renderer
srv, err := sitegen.NewServer(config, renderer)

// Start starts the HTTP server
err := srv.Start()

// StartTLS starts the HTTP server with TLS
err := srv.StartTLS(certFile, keyFile)

// Shutdown gracefully shuts down the server
err := srv.Shutdown()

// Address returns the server address
addr := srv.Address()

// Config returns the server configuration
config := srv.Config()
```

### Renderer

```go
// NewRendererFromFunctions creates a renderer from a map of template names to render functions
renderer := sitegen.NewRendererFromFunctions(map[string]sitegen.RenderFunc{
    "main":   templates.RenderMain,
    "header": templates.RenderHeader,
})

// LoadRenderer attempts to automatically discover and load render functions
renderer, err := sitegen.LoadRenderer(templatePackage)
```

## Advanced Usage

### Custom Renderer

You can create a custom renderer by implementing the `processor.TemplateRenderer` interface:

```go
type TemplateRenderer interface {
    Render(templateName string, w io.Writer, data any) error
}
```

### Processing Individual Files

```go
proc, _ := sitegen.NewProcessor(config, renderer)

// Process a single file
outputPath, content, err := proc.ProcessFile("data/blog/post.json")
if err != nil {
    log.Fatal(err)
}

// Write the output manually
os.WriteFile(outputPath, content, 0644)
```

### Server with Custom Handler

The server uses an internal handler that processes data files on the fly. You can extend this by creating your own HTTP handler that uses the processor:

```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // Use processor to generate content
    _, content, err := proc.ProcessFile("data/index.json")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Write(content)
})
```

