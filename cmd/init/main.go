// Command init creates a new go-hbars project or adds templates and bootstrap to an existing Go module.
//
// Usage:
//
//	init new [path]     Create a new project. path defaults to current directory.
//	init add            Add templates (and optionally bootstrap) to the current module.
//
// Flags for "new":
//   -bootstrap  Include processor/templates, data/, and QuickServer/QuickProcessor main.
//   -module     Module path for go mod init (default: derived from path).
//
// Flags for "add":
//   -bootstrap  Add bootstrap layout (processor/templates, data/, shared/) and example main.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	sub := os.Args[1]
	switch sub {
	case "new":
		runNew(os.Args[2:])
	case "add":
		runAdd(os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		// "init <path>" without "new" -> treat as "init new <path>"
		runNew(os.Args[1:])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `go-hbars init: create a new project or add templates to an existing module.

Usage:
  init new [path]    Create a new project (path defaults to .)
  init add           Add templates and optionally bootstrap to current module

Flags for 'new':
  -bootstrap         Include processor/templates, data/, and QuickServer/QuickProcessor
  -module string      Module path for go mod init (default: from path name)

Flags for 'add':
  -bootstrap         Add bootstrap layout and example main

Examples:
  init new myapp
  init new myapp -bootstrap
  init new . -module myapp -bootstrap
  init add
  init add -bootstrap
`)
}

func runNew(args []string) {
	fs := flag.NewFlagSet("init new", flag.ExitOnError)
	bootstrap := fs.Bool("bootstrap", false, "include processor, data, and QuickServer/QuickProcessor")
	module := fs.String("module", "", "module path for go mod init")
	_ = fs.Parse(args)

	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	// Re-parse so that "init new /path -bootstrap" works (path first, then flags)
	if len(args) >= 2 && args[0] != "" && args[0][0] != '-' {
		path = args[0]
		_ = fs.Parse(args[1:])
	}
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init new: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init new: %v\n", err)
		os.Exit(1)
	}

	if *module == "" {
		*module = filepath.Base(path)
		if *module == "." || *module == "" {
			*module = "myapp"
		}
	}

	// go.mod
	goModPath := filepath.Join(path, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		fmt.Fprintf(os.Stderr, "init new: go.mod already exists in %s\n", path)
		os.Exit(1)
	}
	if err := os.WriteFile(goModPath, []byte("module "+*module+"\n\ngo 1.21\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "init new: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s\n", goModPath)

	if *bootstrap {
		createBootstrapProject(path, *module)
	} else {
		createAPIProject(path, *module)
	}

	runGoGenerate(path)
}

func createAPIProject(root, module string) {
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init new: %v\n", err)
		os.Exit(1)
	}

	genGo := `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates

package templates
`
	writeFile(filepath.Join(templatesDir, "gen.go"), genGo)
	writeFile(filepath.Join(templatesDir, "main.hbs"), "<!DOCTYPE html>\n<html>\n<head><title>{{title}}</title></head>\n<body>\n  {{> header}}\n  <main>{{{content}}}</main>\n  {{> footer}}\n</body>\n</html>\n")
	writeFile(filepath.Join(templatesDir, "header.hbs"), "<header><h1>{{title}}</h1></header>\n")
	writeFile(filepath.Join(templatesDir, "footer.hbs"), "<footer><p>{{note}}</p></footer>\n")

	mainGo := fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"os"

	templates %q
)

func main() {
	dataBytes, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read data: %%v\n", err)
		os.Exit(1)
	}
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "parse data: %%v\n", err)
		os.Exit(1)
	}
	out, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %%v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}
`, module+"/templates")
	writeFile(filepath.Join(root, "main.go"), mainGo)

	sampleData := `{
  "title": "Welcome",
  "content": "Hello, world!",
  "note": "Generated with go-hbars"
}
`
	writeFile(filepath.Join(root, "data.json"), sampleData)
}

func createBootstrapProject(root, module string) {
	templatesDir := filepath.Join(root, "processor", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init new: %v\n", err)
		os.Exit(1)
	}

	genGo := `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
`
	writeFile(filepath.Join(templatesDir, "gen.go"), genGo)
	writeFile(filepath.Join(templatesDir, "main.hbs"), "<!DOCTYPE html>\n<html>\n<head><title>{{title}}</title></head>\n<body>\n  {{> header}}\n  <main>{{{content}}}</main>\n  {{> footer}}\n</body>\n</html>\n")
	writeFile(filepath.Join(templatesDir, "header.hbs"), "<header><h1>{{title}}</h1></header>\n")
	writeFile(filepath.Join(templatesDir, "footer.hbs"), "<footer><p>{{note}}</p></footer>\n")

	dataDir := filepath.Join(root, "data")
	sharedDir := filepath.Join(root, "shared")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init new: %v\n", err)
		os.Exit(1)
	}
	_ = os.MkdirAll(sharedDir, 0755)

	indexJSON := `{
  "_page": {
    "template": "main",
    "output": "index.html"
  },
  "title": "Welcome",
  "content": "Hello, world!",
  "note": "Generated with go-hbars"
}
`
	writeFile(filepath.Join(dataDir, "index.json"), indexJSON)

	mainGo := fmt.Sprintf(`package main

import (
	"flag"
	"log"

	templates %q
)

func main() {
	mode := flag.String("mode", "processor", "processor (static build) or server (HTTP)")
	flag.Parse()
	switch *mode {
	case "server":
		srv, err := templates.NewQuickServer()
		if err != nil {
			log.Fatal(err)
		}
		srv.Config().DataPath = "data"
		srv.Config().Addr = ":8080"
		log.Fatal(srv.Start())
	case "processor":
		proc, err := templates.NewQuickProcessor()
		if err != nil {
			log.Fatal(err)
		}
		proc.Config().DataPath = "data"
		proc.Config().OutputPath = "pages"
		if err := proc.Process(); err != nil {
			log.Fatal(err)
		}
		log.Println("Done. Output in pages/")
	default:
		log.Fatal("use -mode=processor or -mode=server")
	}
}
`, module+"/processor/templates")
	writeFile(filepath.Join(root, "main.go"), mainGo)
}

func runAdd(args []string) {
	fs := flag.NewFlagSet("init add", flag.ExitOnError)
	bootstrap := fs.Bool("bootstrap", false, "add bootstrap layout and example main")
	_ = fs.Parse(args)

	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init add: %v\n", err)
		os.Exit(1)
	}

	modPath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "init add: %v (run from module root)\n", err)
		os.Exit(1)
	}

	if *bootstrap {
		addBootstrap(root, modPath)
	} else {
		addAPI(root, modPath)
	}

	runGoGenerate(root)
}

// runGoGenerate runs go generate ./... and go mod tidy in dir.
func runGoGenerate(dir string) {
	fmt.Printf("Running go generate ./... in %s\n", dir)
	gen := exec.Command("go", "generate", "./...")
	gen.Dir = dir
	gen.Stdout = os.Stdout
	gen.Stderr = os.Stderr
	if err := gen.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "init: go generate failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Running go mod tidy in %s\n", dir)
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	tidy.Stdout = os.Stdout
	tidy.Stderr = os.Stderr
	if err := tidy.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "init: go mod tidy failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

var moduleRe = regexp.MustCompile(`^\s*module\s+(\S+)\s*`)

func readModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := moduleRe.FindStringSubmatch(sc.Text()); m != nil {
			return strings.TrimSpace(m[1]), nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("reading go.mod: %w", err)
	}
	return "", fmt.Errorf("no module directive in go.mod")
}

func addAPI(root, module string) {
	templatesDir := filepath.Join(root, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init add: %v\n", err)
		os.Exit(1)
	}

	genPath := filepath.Join(templatesDir, "gen.go")
	if exists(genPath) {
		fmt.Printf("Skipping %s (already exists)\n", genPath)
	} else {
		genGo := `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates

package templates
`
		writeFile(genPath, genGo)
		fmt.Printf("Created %s\n", genPath)
	}

	for name, body := range map[string]string{
		"main.hbs":   "<!DOCTYPE html>\n<html>\n<head><title>{{title}}</title></head>\n<body>\n  {{> header}}\n  <main>{{{content}}}</main>\n  {{> footer}}\n</body>\n</html>\n",
		"header.hbs": "<header><h1>{{title}}</h1></header>\n",
		"footer.hbs": "<footer><p>{{note}}</p></footer>\n",
	} {
		p := filepath.Join(templatesDir, name)
		if !exists(p) {
			writeFile(p, body)
			fmt.Printf("Created %s\n", p)
		}
	}
}

func addBootstrap(root, module string) {
	templatesDir := filepath.Join(root, "processor", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init add: %v\n", err)
		os.Exit(1)
	}

	genPath := filepath.Join(templatesDir, "gen.go")
	if exists(genPath) {
		fmt.Printf("Skipping %s (already exists)\n", genPath)
	} else {
		genGo := `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
`
		writeFile(genPath, genGo)
		fmt.Printf("Created %s\n", genPath)
	}

	for name, body := range map[string]string{
		"main.hbs":   "<!DOCTYPE html>\n<html>\n<head><title>{{title}}</title></head>\n<body>\n  {{> header}}\n  <main>{{{content}}}</main>\n  {{> footer}}\n</body>\n</html>\n",
		"header.hbs": "<header><h1>{{title}}</h1></header>\n",
		"footer.hbs": "<footer><p>{{note}}</p></footer>\n",
	} {
		p := filepath.Join(templatesDir, name)
		if !exists(p) {
			writeFile(p, body)
			fmt.Printf("Created %s\n", p)
		}
	}

	dataDir := filepath.Join(root, "data")
	sharedDir := filepath.Join(root, "shared")
	_ = os.MkdirAll(dataDir, 0755)
	_ = os.MkdirAll(sharedDir, 0755)

	indexPath := filepath.Join(dataDir, "index.json")
	if !exists(indexPath) {
		writeFile(indexPath, `{
  "_page": {
    "template": "main",
    "output": "index.html"
  },
  "title": "Welcome",
  "content": "Hello, world!",
  "note": "Generated with go-hbars"
}
`)
		fmt.Printf("Created %s\n", indexPath)
	}

	mainPath := filepath.Join(root, "main.go")
	if exists(mainPath) {
		examplePath := filepath.Join(root, "main_hbars_example.go")
		if !exists(examplePath) {
			mainGo := fmt.Sprintf(`// Example bootstrap main. Copy into your main.go or run: go run main_hbars_example.go
package main

import (
	"flag"
	"log"

	templates %q
)

func main() {
	mode := flag.String("mode", "processor", "processor or server")
	flag.Parse()
	switch *mode {
	case "server":
		srv, _ := templates.NewQuickServer()
		srv.Config().DataPath = "data"
		srv.Config().Addr = ":8080"
		log.Fatal(srv.Start())
	case "processor":
		proc, _ := templates.NewQuickProcessor()
		proc.Config().DataPath = "data"
		proc.Config().OutputPath = "pages"
		log.Fatal(proc.Process())
	}
}
`, module+"/processor/templates")
			writeFile(examplePath, mainGo)
			fmt.Printf("Created %s (merge into main.go or run separately)\n", examplePath)
		}
	} else {
		mainGo := fmt.Sprintf(`package main

import (
	"flag"
	"log"

	templates %q
)

func main() {
	mode := flag.String("mode", "processor", "processor or server")
	flag.Parse()
	switch *mode {
	case "server":
		srv, err := templates.NewQuickServer()
		if err != nil {
			log.Fatal(err)
		}
		srv.Config().DataPath = "data"
		srv.Config().Addr = ":8080"
		log.Fatal(srv.Start())
	case "processor":
		proc, err := templates.NewQuickProcessor()
		if err != nil {
			log.Fatal(err)
		}
		proc.Config().DataPath = "data"
		proc.Config().OutputPath = "pages"
		if err := proc.Process(); err != nil {
			log.Fatal(err)
		}
		log.Println("Done. Output in pages/")
	}
}
`, module+"/processor/templates")
		writeFile(mainPath, mainGo)
		fmt.Printf("Created %s\n", mainPath)
	}
}

func writeFile(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
