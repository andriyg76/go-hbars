package sitegen

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/andriyg76/go-hbars/internal/processor"
	"github.com/andriyg76/go-hbars/internal/server"
	"github.com/andriyg76/go-hbars/pkg/renderer"
	"github.com/andriyg76/hexerr"
)

// Server is a semi-static web server that generates pages on the fly.
type Server struct {
	config     *Config
	processor  *processor.Processor
	sharedData map[string]any
	handler    *server.Handler
	httpServer *http.Server
}

// NewServer creates a new server with the given configuration and renderer.
func NewServer(config *Config, r renderer.TemplateRenderer) (*Server, error) {
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
		RootPath:      root,
		DataPath:      config.DataPath,
		SharedPath:    config.SharedPath,
		OutputPath:    "", // Not used for server
	}

	// Load shared data
	sharedPath := filepath.Join(root, config.SharedPath)
	sharedData, err := processor.LoadSharedData(sharedPath)
	if err != nil {
		return nil, hexerr.Wrap(err, "failed to load shared data")
	}

	// Create processor
	proc := processor.NewProcessor(procConfig, r)

	// Create handler
	staticDir := ""
	if config.StaticDir != "" {
		staticDir = filepath.Join(root, config.StaticDir)
	}
	handler := server.NewHandler(proc, sharedData, staticDir)

	// Create HTTP server
	addr := config.Addr
	if addr == "" {
		addr = ":8080"
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return &Server{
		config:     config,
		processor:  proc,
		sharedData: sharedData,
		handler:    handler,
		httpServer: httpServer,
	}, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// StartTLS starts the HTTP server with TLS.
func (s *Server) StartTLS(certFile, keyFile string) error {
	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(nil)
	}
	return nil
}

// Address returns the server address.
func (s *Server) Address() string {
	return s.httpServer.Addr
}

// Config returns the server configuration.
func (s *Server) Config() *Config {
	return s.config
}
