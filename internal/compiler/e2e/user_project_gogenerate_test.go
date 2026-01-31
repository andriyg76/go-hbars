package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_UserProject_GoGenerate_CompatShowcase runs a user-style project that uses
// go-hbars from GitHub (no replace): go:generate with go run .../cmd/hbc@latest,
// main reads compat and showcase data and renders both.
func TestE2E_UserProject_GoGenerate_CompatShowcase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	root := repoRoot(t)
	examplesCompat := filepath.Join(root, "examples", "compat")
	examplesShowcase := filepath.Join(root, "examples", "showcase")

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

	compatMain, _ := os.ReadFile(filepath.Join(examplesCompat, "templates", "main.hbs"))
	compatMainStr := strings.Replace(string(compatMain), "{{> footer note=\"thanks\"}}", "{{> compat_footer note=\"thanks\"}}", 1)
	writeFile("templates/compat.hbs", compatMainStr)
	copyFile("templates/showcase.hbs", filepath.Join(examplesShowcase, "templates", "main.hbs"))
	copyFile("templates/compat_footer.hbs", filepath.Join(examplesCompat, "templates", "footer.hbs"))
	copyFile("templates/header.hbs", filepath.Join(examplesShowcase, "templates", "header.hbs"))
	copyFile("templates/footer.hbs", filepath.Join(examplesShowcase, "templates", "footer.hbs"))
	copyFile("templates/userCard.hbs", filepath.Join(examplesCompat, "templates", "userCard.hbs"))
	copyFile("templates/orderRow.hbs", filepath.Join(examplesShowcase, "templates", "orderRow.hbs"))

	// go:generate without @latest so replace in go.mod uses local hbc.
	writeFile("templates/gen.go", `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc -in . -out ./templates_gen.go -pkg templates

package templates
`)
	// Import hbc so go mod tidy keeps the dependency and adds transitive deps to go.sum.
	writeFile("tools.go", `//go:build tools

package main

import _ "github.com/andriyg76/go-hbars/cmd/hbc"
`)

	// Use local go-hbars so go:generate runs the local hbc (not @latest from network).
	writeFile("go.mod", "module test-e2e-api\n\ngo 1.24\n\nreplace github.com/andriyg76/go-hbars => "+filepath.ToSlash(root)+"\n")

	if err := os.MkdirAll(filepath.Join(tmpDir, "data"), 0755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	copyFile("data/compat.json", filepath.Join(examplesCompat, "data.json"))
	copyFile("data/showcase.json", filepath.Join(examplesShowcase, "data.json"))
	copyFile("expected.txt", filepath.Join(examplesCompat, "expected.txt"))

	writeFile("main.go", `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	templates "test-e2e-api/templates"
)

func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, " \t")
		line = regexp.MustCompile(" +").ReplaceAllString(line, " ")
		line = regexp.MustCompile(":([^\\s])").ReplaceAllString(line, ": $1")
		lines[i] = line
	}
	s = strings.Join(lines, "\n")
	s = regexp.MustCompile("\\n{2,}").ReplaceAllString(s, "\n")
	return strings.Trim(s, "\n")
}

func normalizeMapKeyOrder(s string) string {
	lines := strings.Split(s, "\n")
	var i int
	for i < len(lines) {
		var run []int
		for i < len(lines) {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed == "" || !strings.Contains(trimmed, "=") {
				i++
				continue
			}
			if strings.HasPrefix(lines[i], " ") && len(trimmed) > 0 {
				run = append(run, i)
				i++
			} else {
				break
			}
		}
		if len(run) > 1 {
			group := make([]string, len(run))
			for j, idx := range run {
				group[j] = lines[idx]
			}
			sort.Strings(group)
			for j, idx := range run {
				lines[idx] = group[j]
			}
		}
		if len(run) == 0 {
			i++
		}
	}
	return strings.Join(lines, "\n")
}

func main() {
	compatBytes, err := os.ReadFile("data/compat.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read compat: %v\n", err)
		os.Exit(1)
	}
	var compatData map[string]any
	if err := json.Unmarshal(compatBytes, &compatData); err != nil {
		fmt.Fprintf(os.Stderr, "parse compat: %v\n", err)
		os.Exit(1)
	}
	showcaseBytes, err := os.ReadFile("data/showcase.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read showcase: %v\n", err)
		os.Exit(1)
	}
	var showcaseData map[string]any
	if err := json.Unmarshal(showcaseBytes, &showcaseData); err != nil {
		fmt.Fprintf(os.Stderr, "parse showcase: %v\n", err)
		os.Exit(1)
	}

	compatOut, err := templates.RenderCompatString(templates.CompatContextFromMap(compatData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render compat: %v\n", err)
		os.Exit(1)
	}
	showcaseOut, err := templates.RenderShowcaseString(templates.ShowcaseContextFromMap(showcaseData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render showcase: %v\n", err)
		os.Exit(1)
	}

	expectedBytes, err := os.ReadFile("expected.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read expected: %v\n", err)
		os.Exit(1)
	}
	expected := string(expectedBytes)
	compatOut = strings.ReplaceAll(compatOut, "\r\n", "\n")
	expected = strings.ReplaceAll(expected, "\r\n", "\n")
	compatOut = normalizeWhitespace(compatOut)
	expected = normalizeWhitespace(expected)
	compatOut = normalizeMapKeyOrder(compatOut)
	expected = normalizeMapKeyOrder(expected)
	if compatOut != expected {
		fmt.Fprintf(os.Stderr, "compat output mismatch\n")
		os.Exit(1)
	}
	if strings.TrimSpace(showcaseOut) == "" {
		fmt.Fprintf(os.Stderr, "showcase output empty\n")
		os.Exit(1)
	}
	fmt.Println("OK")
}
`)

	// Get module so replace applies; tools.go imports hbc so tidy keeps it and adds transitive deps.
	cmd := exec.Command("go", "get", "github.com/andriyg76/go-hbars")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go get: %v\n%s", err, out)
	}
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
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

	cmd = exec.Command("go", "run", ".")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "OK") {
		t.Fatalf("expected OK, got:\n%s", output)
	}
}
