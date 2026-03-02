// Build page.html, index.html, and news.html from templates and data.
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

// dataFiles lists JSON files; each is rendered with the matching template to the corresponding HTML.
var dataFiles = []string{"index.json", "firstpage.json", "news.json"}

func main() {
	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}
	dataDir := filepath.Join(baseDir, "data")

	for _, dataFile := range dataFiles {
		data, err := loadData(filepath.Join(dataDir, dataFile))
		if err != nil {
			panic(err)
		}
		var html string
		switch dataFile {
		case "index.json":
			html, err = templates.RenderPageString(templates.PageContextFromMap(data))
		case "firstpage.json":
			html, err = templates.RenderFirstpageString(templates.FirstpageContextFromMap(data))
		case "news.json":
			html, err = templates.RenderNewsString(templates.NewsContextFromMap(data))
		default:
			panic("unknown data file: " + dataFile)
		}
		if err != nil {
			panic(err)
		}
		outPath := filepath.Join(baseDir, outputName(dataFile))
		if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
			panic(err)
		}
	}
}

func outputName(dataFile string) string {
	switch dataFile {
	case "index.json":
		return "page.html"
	case "firstpage.json":
		return "index.html"
	case "news.json":
		return "news.html"
	default:
		return "out.html"
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
