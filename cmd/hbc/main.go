// Command hbc compiles Handlebars templates to Go source.
//
// Usage:
//
//	hbc -in <dir|file> -out <file.go> -pkg <package>
//
// Phased compilation (observe intermediate results):
//
//	hbc -in templates -max-phase 1 -phase1-output phase1_ast.json
//	hbc -in templates -max-phase 2 -phase2a-output phase2a.json -phase2b-output phase2b.json
//	hbc -in templates -max-phase 4 -output gen/templates_gen.go
//
// Logging:
//
//	-verbose  Debug-level compiler messages
//	-trace    Trace-level (finest) compiler messages
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andriyg76/glog"
	"github.com/andriyg76/go-hbars/internal/compiler"
	"github.com/andriyg76/hexerr"
)

func main() {
	hexerr.SetFilterPrefixes("github.com/andriyg76/go-hbars")

	var (
		inDir       = flag.String("in", "", "input directory or file (templates); required")
		outPath     = flag.String("out", "", "output .go file path (default: stdout when max-phase=4)")
		pkg         = flag.String("pkg", "templates", "Go package name for generated code")
		maxPhase    = flag.Int("max-phase", 4, "maximum compilation phase (1=parse, 2=2a, 3=2b, 4=IR, 5=Go); compilation runs from 1 to max-phase")
		phase1Out   = flag.String("phase1-output", "", "write phase 1 result (AST) to this JSON file")
		phase2aOut  = flag.String("phase2a-output", "", "write phase 2a result (partial types) to this JSON file")
		phase2bOut  = flag.String("phase2b-output", "", "write phase 2b result (type trees) to this JSON file")
		phase3Out   = flag.String("phase3-output", "", "write phase 3 result (IR) to this JSON file")
		verbose     = flag.Bool("verbose", false, "enable debug-level compiler log")
		trace       = flag.Bool("trace", false, "enable trace-level compiler log")
		bootstrap   = flag.Bool("bootstrap", false, "generate bootstrap (NewRenderer, NewQuickProcessor, NewQuickServer)")
		genVersion  = flag.String("generator-version", "", "emit Generator version comment in output")
		runtimeImp  = flag.String("runtime", "", "runtime import path (default: github.com/andriyg76/go-hbars/runtime)")
	)
	flag.Parse()

	if *inDir == "" {
		fmt.Fprintf(os.Stderr, "hbc: -in is required\n")
		flag.Usage()
		os.Exit(2)
	}

	// Control default log output level via glog: -verbose => DEBUG, -trace => TRACE.
	// The compiler (and any code) then uses the default logger (glog.Info, glog.Debug, glog.Trace).
	if *trace {
		glog.SetLevel(glog.TRACE)
	} else if *verbose {
		glog.SetLevel(glog.DEBUG)
	}
	compilerLog := &glogDefaultLog{}

	templates, templateFiles, err := loadTemplates(*inDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hbc: %v\n", err)
		os.Exit(1)
	}
	if len(templates) == 0 {
		fmt.Fprintf(os.Stderr, "hbc: no .hbs templates found in %s\n", *inDir)
		os.Exit(1)
	}

	opts := compiler.Options{
		PackageName:       *pkg,
		RuntimeImport:     *runtimeImp,
		GenerateBootstrap: *bootstrap,
		GeneratorVersion:  *genVersion,
		TemplateFiles:     templateFiles,
		Log:               compilerLog,
	}
	opts.MaxPhase = compiler.MaxPhase(*maxPhase)
	if opts.MaxPhase < 1 || opts.MaxPhase > 5 {
		fmt.Fprintf(os.Stderr, "hbc: -max-phase must be 1..5\n")
		os.Exit(2)
	}
	opts.Phase1Output = *phase1Out
	opts.Phase2aOutput = *phase2aOut
	opts.Phase2bOutput = *phase2bOut
	opts.Phase3Output = *phase3Out
	opts.Phase4Output = *outPath

	out, err := compiler.CompileTemplates(templates, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hbc: %v\n", err)
		os.Exit(1)
	}

	if len(out) > 0 && *outPath == "" {
		os.Stdout.Write(out)
	}
	glog.Info("Done.")
}

// glogDefaultLog forwards to glog package-level functions (default logger).
// hbc sets the level with glog.SetLevel so the default logger controls what is emitted.
type glogDefaultLog struct{}

func (glogDefaultLog) Info(format string, args ...any)  { glog.Info(format, args...) }
func (glogDefaultLog) Debug(format string, args ...any) { glog.Debug(format, args...) }
func (glogDefaultLog) Trace(format string, args ...any) { glog.Trace(format, args...) }

// loadTemplates loads all .hbs files from dir (or a single file) and returns map[name]content and map[name]filepath.
func loadTemplates(in string) (map[string]string, map[string]string, error) {
	info, err := os.Stat(in)
	if err != nil {
		return nil, nil, err
	}
	templates := make(map[string]string)
	templateFiles := make(map[string]string)
	if info.IsDir() {
		entries, err := os.ReadDir(in)
		if err != nil {
			return nil, nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !strings.HasSuffix(e.Name(), ".hbs") {
				continue
			}
			path := filepath.Join(in, e.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, hexerr.Wrapf(err, "read %s", path)
			}
			name := strings.TrimSuffix(e.Name(), ".hbs")
			if name == "" {
				continue
			}
			templates[name] = string(content)
			templateFiles[name] = path
		}
		return templates, templateFiles, nil
	}
	if !strings.HasSuffix(in, ".hbs") {
		return nil, nil, hexerr.New("single file input must be .hbs")
	}
	content, err := os.ReadFile(in)
	if err != nil {
		return nil, nil, err
	}
	name := strings.TrimSuffix(filepath.Base(in), ".hbs")
	templates[name] = string(content)
	templateFiles[name] = in
	return templates, templateFiles, nil
}
