// Build index.html and news.html from templates and data.
//
// 1. Generate Go code from templates:
//    go run github.com/andriyg76/go-hbars/cmd/hbc -in templates -out gen/templates_gen.go -pkg templates
//
// 2. Run this program:
//    go run .
package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	templates "github.com/andriyg76/go-hbars/examples/partials-inheritance/gen"
)

func main() {
	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}
	dataDir := filepath.Join(baseDir, "data")

	// Load index data and render to index.html
	indexData, err := loadData(filepath.Join(dataDir, "index.json"))
	if err != nil {
		panic(err)
	}
	indexHTML, err := templates.RenderDefaultString(templates.DefaultContextFromMap(indexData))
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "index.html"), []byte(indexHTML), 0644); err != nil {
		panic(err)
	}

	// Load news data and render to news.html
	newsData, err := loadData(filepath.Join(dataDir, "news.json"))
	if err != nil {
		panic(err)
	}
	newsHTML, err := templates.RenderNewsString(templates.NewsContextFromMap(newsData))
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "news.html"), []byte(newsHTML), 0644); err != nil {
		panic(err)
	}
}

func loadData(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	// Remove _page so it is not passed to templates
	delete(out, "_page")
	return out, nil
}
