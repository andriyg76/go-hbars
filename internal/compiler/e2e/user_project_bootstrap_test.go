package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_UserProject_Bootstrap_ServerAndProcessor runs a user-style project with
// -bootstrap: go:generate with hbc from GitHub, data files with _page, then
// runs NewQuickProcessor (static pages).
func TestE2E_UserProject_Bootstrap_ServerAndProcessor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	root := repoRoot(t)
	examplesCompat := filepath.Join(root, "examples", "compat")

	tmpDir := t.TempDir()
	writeFile := func(path, content string) {
		full := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	copyFile := func(dest, src string) {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		writeFile(dest, string(data))
	}

	procTpl := filepath.Join(tmpDir, "processor", "templates")
	if err := os.MkdirAll(procTpl, 0755); err != nil {
		t.Fatalf("mkdir processor/templates: %v", err)
	}
	compatMain, _ := os.ReadFile(filepath.Join(examplesCompat, "templates", "main.hbs"))
	compatMainStr := strings.Replace(string(compatMain), "{{> footer note=\"thanks\"}}", "{{> compat_footer note=\"thanks\"}}", 1)
	writeFile("processor/templates/compat.hbs", compatMainStr)
	copyFile("processor/templates/compat_footer.hbs", filepath.Join(examplesCompat, "templates", "footer.hbs"))
	copyFile("processor/templates/userCard.hbs", filepath.Join(examplesCompat, "templates", "userCard.hbs"))

	writeFile("processor/templates/gen.go", `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
`)

	repoRootPath := strings.ReplaceAll(root, "\\", "/")
	writeFile("go.mod", `module test-e2e-bootstrap

go 1.24

require github.com/andriyg76/go-hbars v0.0.0
replace github.com/andriyg76/go-hbars => `+repoRootPath+`
`)

	if err := os.MkdirAll(filepath.Join(tmpDir, "data"), 0755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "shared"), 0755); err != nil {
		t.Fatalf("mkdir shared: %v", err)
	}
	compatBytes, _ := os.ReadFile(filepath.Join(examplesCompat, "data.json"))
	var compatMap map[string]any
	json.Unmarshal(compatBytes, &compatMap)
	compatMap["_page"] = map[string]string{"template": "compat", "output": "compat.html"}
	b, _ := json.MarshalIndent(compatMap, "", "  ")
	writeFile("data/compat.json", string(b))
	indexMap := make(map[string]any)
	json.Unmarshal(compatBytes, &indexMap)
	indexMap["_page"] = map[string]string{"template": "compat", "output": "index.html"}
	b, _ = json.MarshalIndent(indexMap, "", "  ")
	writeFile("data/index.json", string(b))

	writeFile("main.go", `package main

import (
	"flag"
	"log"

	templates "test-e2e-bootstrap/processor/templates"
)

func main() {
	mode := flag.String("mode", "", "server or processor")
	flag.Parse()
	switch *mode {
	case "server":
		srv, err := templates.NewQuickServer()
		if err != nil {
			log.Fatal(err)
		}
		srv.Config().Addr = ":19999"
		log.Fatal(srv.Start())
	case "processor":
		proc, err := templates.NewQuickProcessor()
		if err != nil {
			log.Fatal(err)
		}
		if err := proc.Process(); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("use -mode=server or -mode=processor")
	}
}
`)

	cmd := exec.Command("go", "get", "github.com/andriyg76/go-hbars/...")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go get go-hbars/...: %v\n%s", err, out)
	}
	cmd = exec.Command("go", "generate", "./...")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go generate: %v\n%s", err, out)
	}
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	procCmd := exec.Command("go", "run", ".", "-mode=processor")
	procCmd.Dir = tmpDir
	if out, err := procCmd.CombinedOutput(); err != nil {
		t.Fatalf("go run -mode=processor: %v\n%s", err, out)
	}
	compatHTML := filepath.Join(tmpDir, "pages", "compat.html")
	compatContent, err := os.ReadFile(compatHTML)
	if err != nil {
		t.Fatalf("read compat.html: %v", err)
	}
	if len(compatContent) == 0 {
		t.Fatalf("compat.html empty")
	}
	if !strings.Contains(string(compatContent), "Handlebars Compat") {
		t.Fatalf("compat.html missing expected content")
	}
}
