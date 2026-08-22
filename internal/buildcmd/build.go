package buildcmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

const (
	modulePath        = "github.com/andriyg76/go-hbars"
	requiredGoVersion = "1.24"
)

type Options struct {
	RootPath      string
	TemplatesPath string
	DataPath      string
	SharedPath    string
	OutputPath    string
	GoBinary      string
	ModuleVersion string
	ModuleReplace string
}

type Result struct {
	TemplateCompile time.Duration
	GoBuild         time.Duration
	Render          time.Duration
}

func RunCLI(args []string, output io.Writer) error {
	if output == nil {
		output = io.Discard
	}
	opts := Options{}
	flags := flag.NewFlagSet("hbc build", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&opts.RootPath, "root", "", "site root (default: current directory)")
	flags.StringVar(&opts.TemplatesPath, "templates-path", ".processor/templates", "templates path, relative to root unless absolute")
	flags.StringVar(&opts.DataPath, "data-path", "data", "data path, relative to root unless absolute")
	flags.StringVar(&opts.SharedPath, "shared-path", "shared", "shared data path, relative to root unless absolute")
	flags.StringVar(&opts.OutputPath, "output-path", "pages", "generated site output path, relative to root unless absolute")
	flags.StringVar(&opts.GoBinary, "go", "go", "Go executable used to build the generated renderer")
	flags.StringVar(&opts.ModuleVersion, "go-hbars-version", defaultModuleVersion(), "go-hbars runtime version used by the generated renderer")
	flags.StringVar(&opts.ModuleReplace, "go-hbars-replace", "", "optional local go-hbars module directory for development")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}

	result, err := Run(opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(output, "Templates compiled in %s\n", result.TemplateCompile)
	fmt.Fprintf(output, "Renderer built in %s\n", result.GoBuild)
	fmt.Fprintf(output, "Site generated in %s\n", result.Render)
	return nil
}

func Run(opts Options) (Result, error) {
	var result Result
	resolved, err := resolveOptions(opts)
	if err != nil {
		return result, err
	}

	templates, templateFiles, err := compiler.LoadTemplates(resolved.TemplatesPath)
	if err != nil {
		return result, fmt.Errorf("load templates: %w", err)
	}
	if len(templates) == 0 {
		return result, fmt.Errorf("no .hbs templates found in %s", resolved.TemplatesPath)
	}

	directory, err := os.MkdirTemp("", "go-hbars-build-")
	if err != nil {
		return result, fmt.Errorf("create build directory: %w", err)
	}
	defer os.RemoveAll(directory)

	templatesDir := filepath.Join(directory, "templates")
	mainDir := filepath.Join(directory, "cmd", "render")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		return result, fmt.Errorf("create generated templates directory: %w", err)
	}
	if err := os.MkdirAll(mainDir, 0o755); err != nil {
		return result, fmt.Errorf("create renderer command directory: %w", err)
	}

	helperRegistry := helpers.Registry()
	compilerHelpers := make(map[string]compiler.HelperRef, len(helperRegistry))
	for name, ref := range helperRegistry {
		compilerHelpers[name] = compiler.HelperRef{ImportPath: ref.ImportPath, Ident: ref.Ident}
	}
	compileStarted := time.Now()
	generated, err := compiler.CompileTemplates(templates, compiler.Options{
		PackageName:       "templates",
		Helpers:           compilerHelpers,
		GenerateBootstrap: true,
		TemplateFiles:     templateFiles,
	})
	if err != nil {
		return result, fmt.Errorf("compile templates: %w", err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "templates_gen.go"), generated, 0o644); err != nil {
		return result, fmt.Errorf("write generated templates: %w", err)
	}
	result.TemplateCompile = time.Since(compileStarted)

	goMod, err := moduleFile(resolved.ModuleVersion, resolved.ModuleReplace)
	if err != nil {
		return result, err
	}
	if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte(goMod), 0o644); err != nil {
		return result, fmt.Errorf("write renderer go.mod: %w", err)
	}
	if err := os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(runnerMain), 0o644); err != nil {
		return result, fmt.Errorf("write renderer main: %w", err)
	}

	executable := filepath.Join(directory, "site-render")
	buildStarted := time.Now()
	if err := runCommand(directory, resolved.GoBinary, "build", "-mod=mod", "-o", executable, "./cmd/render"); err != nil {
		return result, fmt.Errorf("build renderer: %w", err)
	}
	result.GoBuild = time.Since(buildStarted)

	renderStarted := time.Now()
	if err := runCommand("", executable,
		"-root", resolved.RootPath,
		"-data-path", resolved.DataPath,
		"-shared-path", resolved.SharedPath,
		"-output-path", resolved.OutputPath,
	); err != nil {
		return result, fmt.Errorf("generate site: %w", err)
	}
	result.Render = time.Since(renderStarted)
	return result, nil
}

func resolveOptions(opts Options) (Options, error) {
	if strings.TrimSpace(opts.RootPath) == "" {
		root, err := os.Getwd()
		if err != nil {
			return Options{}, fmt.Errorf("get current directory: %w", err)
		}
		opts.RootPath = root
	}
	root, err := filepath.Abs(opts.RootPath)
	if err != nil {
		return Options{}, fmt.Errorf("resolve root: %w", err)
	}
	opts.RootPath = root

	if strings.TrimSpace(opts.TemplatesPath) == "" {
		opts.TemplatesPath = ".processor/templates"
	}
	opts.TemplatesPath = resolvePath(opts.TemplatesPath, root)
	if strings.TrimSpace(opts.DataPath) == "" {
		opts.DataPath = "data"
	}
	if strings.TrimSpace(opts.SharedPath) == "" {
		opts.SharedPath = "shared"
	}
	if strings.TrimSpace(opts.OutputPath) == "" {
		opts.OutputPath = "pages"
	}
	if strings.TrimSpace(opts.GoBinary) == "" {
		opts.GoBinary = "go"
	}
	if _, err := exec.LookPath(opts.GoBinary); err != nil {
		return Options{}, fmt.Errorf("Go executable %q not found: %w", opts.GoBinary, err)
	}
	if strings.TrimSpace(opts.ModuleVersion) == "" {
		opts.ModuleVersion = defaultModuleVersion()
	}
	if strings.ContainsAny(opts.ModuleVersion, " \t\r\n") {
		return Options{}, fmt.Errorf("invalid go-hbars version %q", opts.ModuleVersion)
	}
	return opts, nil
}

func resolvePath(path, root string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}

func defaultModuleVersion() string {
	return "v" + strings.TrimPrefix(compiler.Version, "v")
}

func moduleFile(version, replace string) (string, error) {
	goMod := fmt.Sprintf("module teatr.local/site-renderer\n\ngo %s\n\nrequire %s %s\n", requiredGoVersion, modulePath, version)
	if strings.TrimSpace(replace) == "" {
		return goMod, nil
	}
	replacePath, err := filepath.Abs(replace)
	if err != nil {
		return "", fmt.Errorf("resolve go-hbars replacement: %w", err)
	}
	return goMod + fmt.Sprintf("\nreplace %s => %s\n", modulePath, filepath.ToSlash(replacePath)), nil
}

func runCommand(directory, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}
	return nil
}

const runnerMain = `package main

import (
	"flag"
	"log"

	"github.com/andriyg76/go-hbars/pkg/sitegen"
	"teatr.local/site-renderer/templates"
)

func main() {
	root := flag.String("root", "", "site root")
	dataPath := flag.String("data-path", "data", "data path")
	sharedPath := flag.String("shared-path", "shared", "shared data path")
	outputPath := flag.String("output-path", "pages", "generated site output path")
	flag.Parse()
	if *root == "" {
		log.Fatal("root is required")
	}

	config := sitegen.DefaultConfig()
	config.RootPath = *root
	config.DataPath = *dataPath
	config.SharedPath = *sharedPath
	config.OutputPath = *outputPath
	processor, err := sitegen.NewProcessor(config, templates.NewRenderer())
	if err != nil {
		log.Fatal(err)
	}
	if err := processor.Process(); err != nil {
		log.Fatal(err)
	}
}
`
