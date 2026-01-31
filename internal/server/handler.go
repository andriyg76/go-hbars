package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/andriyg76/glog"
	"github.com/andriyg76/go-hbars/internal/processor"
)

var serverLog = glog.Create(glog.INFO)

// Handler handles HTTP requests for the semi-static server.
type Handler struct {
	processor  *processor.Processor
	sharedData map[string]any
	staticDir  string
}

// NewHandler creates a new HTTP handler.
func NewHandler(proc *processor.Processor, sharedData map[string]any, staticDir string) *Handler {
	return &Handler{
		processor:  proc,
		sharedData: sharedData,
		staticDir:  staticDir,
	}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle static files
	if h.staticDir != "" && h.handleStatic(w, r) {
		return
	}

	// Handle data file requests
	if h.handleDataFile(w, r) {
		return
	}

	// 404
	http.NotFound(w, r)
}

// handleStatic serves static files if they exist.
func (h *Handler) handleStatic(w http.ResponseWriter, r *http.Request) bool {
	if h.staticDir == "" {
		return false
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	staticPath := filepath.Join(h.staticDir, path)
	info, err := os.Stat(staticPath)
	if err != nil || info.IsDir() {
		return false
	}

	http.ServeFile(w, r, staticPath)
	return true
}

// handleDataFile handles requests for data files.
func (h *Handler) handleDataFile(w http.ResponseWriter, r *http.Request) bool {
	// Map URL path to data file
	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	if urlPath == "" {
		urlPath = "index"
	}

	// Try to find matching data file
	dataPath := h.findDataFile(urlPath)
	if dataPath == "" {
		return false
	}

	// Process the file
	outputPath, content, err := h.processor.ProcessFile(dataPath, h.sharedData)
	if err != nil {
		serverLog.Error("Failed to process file: %+v", err)
		http.Error(w, fmt.Sprintf("Failed to process file: %v", err), http.StatusInternalServerError)
		return true
	}
	if outputPath == "" {
		return false
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
	return true
}

// findDataFile finds a data file matching the URL path.
func (h *Handler) findDataFile(urlPath string) string {
	config := h.processor.Config()
	if config == nil {
		return ""
	}

	// Resolve data path
	dataPath := config.DataPath
	if !filepath.IsAbs(dataPath) && config.RootPath != "" {
		dataPath = filepath.Join(config.RootPath, dataPath)
	}

	// Try different extensions
	extensions := []string{".json", ".yaml", ".yml", ".toml"}
	basePath := filepath.Join(dataPath, urlPath)

	for _, ext := range extensions {
		filePath := basePath + ext
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			return filePath
		}
	}

	// Try index file in directory
	for _, ext := range extensions {
		filePath := filepath.Join(basePath, "index"+ext)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			return filePath
		}
	}

	return ""
}
